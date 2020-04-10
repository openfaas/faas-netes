// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/openfaas/faas-netes/handlers"
	"github.com/openfaas/faas-netes/k8s"
	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	informers "github.com/openfaas/faas-netes/pkg/client/informers/externalversions"
	"github.com/openfaas/faas-netes/pkg/controller"
	"github.com/openfaas/faas-netes/pkg/server"
	"github.com/openfaas/faas-netes/pkg/signals"
	"github.com/openfaas/faas-netes/types"
	version "github.com/openfaas/faas-netes/version"
	faasProvider "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"
	providertypes "github.com/openfaas/faas-provider/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	glog "k8s.io/klog"

	// required to authenticate against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// required for updating and validating the CRD clientset
	_ "k8s.io/code-generator/cmd/client-gen/generators"
)

var pullPolicyOptions = map[string]bool{
	"Always":       true,
	"IfNotPresent": true,
	"Never":        true,
}

func main() {
	var kubeconfig string
	var masterURL string
	var operator bool

	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	flag.BoolVar(&operator, "operator", false, "Use the operator mode instead of faas-netes")
	flag.Parse()

	if !operator {
		runController(kubeconfig, masterURL)
	} else {
		runOperator(kubeconfig, masterURL)
	}
}

// runController runs the faas-netes imperative controller
func runController(kubeconfig, masterURL string) {
	clientCmdConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(clientCmdConfig)
	if err != nil {
		log.Fatalf("Error building Kubernetes clientset: %s", err.Error())

	}

	functionNamespace := "default"

	if namespace, exists := os.LookupEnv("function_namespace"); exists {
		functionNamespace = namespace
	}

	readConfig := types.ReadConfig{}
	osEnv := providertypes.OsEnv{}
	cfg, err := readConfig.Read(osEnv)
	if err != nil {
		log.Fatalf("Error reading config: %s", err.Error())
	}

	log.Printf("HTTP Read Timeout: %s\n", cfg.FaaSConfig.GetReadTimeout())
	log.Printf("HTTP Write Timeout: %s\n", cfg.FaaSConfig.WriteTimeout)
	log.Printf("HTTPProbe: %v\n", cfg.HTTPProbe)
	log.Printf("SetNonRootUser: %v\n", cfg.SetNonRootUser)

	sha, release := version.GetReleaseInfo()
	log.Printf("Starting operator. Version: %s\tcommit: %s", release, sha)

	deployConfig := k8s.DeploymentConfig{
		RuntimeHTTPPort: 8080,
		HTTPProbe:       cfg.HTTPProbe,
		SetNonRootUser:  cfg.SetNonRootUser,
		ReadinessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(cfg.ReadinessProbeInitialDelaySeconds),
			TimeoutSeconds:      int32(cfg.ReadinessProbeTimeoutSeconds),
			PeriodSeconds:       int32(cfg.ReadinessProbePeriodSeconds),
		},
		LivenessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(cfg.LivenessProbeInitialDelaySeconds),
			TimeoutSeconds:      int32(cfg.LivenessProbeTimeoutSeconds),
			PeriodSeconds:       int32(cfg.LivenessProbePeriodSeconds),
		},
		ImagePullPolicy: cfg.ImagePullPolicy,
	}

	factory := k8s.NewFunctionFactory(clientset, deployConfig)

	defaultResync := time.Second * 5
	kubeInformerOpt := kubeinformers.WithNamespace(functionNamespace)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clientset, defaultResync, kubeInformerOpt)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	endpointsInformer := kubeInformerFactory.Core().V1().Endpoints()
	go kubeInformerFactory.Start(stopCh)
	lister := endpointsInformer.Lister()

	functionLookup := k8s.NewFunctionLookup(functionNamespace, lister)

	bootstrapHandlers := providertypes.FaaSHandlers{
		FunctionProxy:        proxy.NewHandlerFunc(cfg.FaaSConfig, functionLookup),
		DeleteHandler:        handlers.MakeDeleteHandler(functionNamespace, clientset),
		DeployHandler:        handlers.MakeDeployHandler(functionNamespace, factory),
		FunctionReader:       handlers.MakeFunctionReader(functionNamespace, clientset),
		ReplicaReader:        handlers.MakeReplicaReader(functionNamespace, clientset),
		ReplicaUpdater:       handlers.MakeReplicaUpdater(functionNamespace, clientset),
		UpdateHandler:        handlers.MakeUpdateHandler(functionNamespace, factory),
		HealthHandler:        handlers.MakeHealthHandler(),
		InfoHandler:          handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommit),
		SecretHandler:        handlers.MakeSecretHandler(functionNamespace, clientset),
		LogHandler:           logs.NewLogHandlerFunc(k8s.NewLogRequestor(clientset, functionNamespace), cfg.FaaSConfig.WriteTimeout),
		ListNamespaceHandler: handlers.MakeNamespacesLister(functionNamespace, clientset),
	}

	faasProvider.Serve(&bootstrapHandlers, &cfg.FaaSConfig)
}

// runOperator runs the CRD Operator
func runOperator(kubeconfig, masterURL string) {

	setupLogging()

	sha, release := version.GetReleaseInfo()
	glog.Infof("Starting operator. Version: %s\tcommit: %s", release, sha)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building Kubernetes clientset: %s", err.Error())
	}

	faasClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building OpenFaaS clientset: %s", err.Error())
	}

	readConfig := types.ReadConfig{}
	osEnv := providertypes.OsEnv{}
	config, err := readConfig.Read(osEnv)

	if err != nil {
		panic(err)
	}

	deployConfig := k8s.DeploymentConfig{
		RuntimeHTTPPort: 8080,
		HTTPProbe:       config.HTTPProbe,
		SetNonRootUser:  config.SetNonRootUser,
		ReadinessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(config.ReadinessProbeInitialDelaySeconds),
			TimeoutSeconds:      int32(config.ReadinessProbeTimeoutSeconds),
			PeriodSeconds:       int32(config.ReadinessProbePeriodSeconds),
		},
		LivenessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(config.LivenessProbeInitialDelaySeconds),
			TimeoutSeconds:      int32(config.LivenessProbeTimeoutSeconds),
			PeriodSeconds:       int32(config.LivenessProbePeriodSeconds),
		},
		ImagePullPolicy: config.ImagePullPolicy,
	}

	factory := controller.NewFunctionFactory(kubeClient, deployConfig)

	functionNamespace := "openfaas-fn"
	if namespace, exists := os.LookupEnv("function_namespace"); exists {
		functionNamespace = namespace
	}

	if !pullPolicyOptions[config.ImagePullPolicy] {
		glog.Fatalf("Invalid image_pull_policy configured: %s", config.ImagePullPolicy)
	}

	// the sync interval does not affect the scale to/from zero feature
	// auto-scaling is does via the HTTP API that acts on the deployment Spec.Replicas
	defaultResync := time.Minute * 5

	kubeInformerOpt := kubeinformers.WithNamespace(functionNamespace)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResync, kubeInformerOpt)

	faasInformerOpt := informers.WithNamespace(functionNamespace)
	faasInformerFactory := informers.NewSharedInformerFactoryWithOptions(faasClient, defaultResync, faasInformerOpt)

	endpointsInformer := kubeInformerFactory.Core().V1().Endpoints()
	deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()

	log.Printf("Waiting for cache sync in main")
	kubeInformerFactory.WaitForCacheSync(stopCh)
	log.Printf("Cache sync done")

	ctrl := controller.NewController(
		kubeClient,
		faasClient,
		kubeInformerFactory,
		faasInformerFactory,
		factory,
	)

	srv := server.New(faasClient, kubeClient, endpointsInformer, deploymentInformer)

	go faasInformerFactory.Start(stopCh)
	go kubeInformerFactory.Start(stopCh)

	go srv.Start()
	if err = ctrl.Run(1, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

func setupLogging() {
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	glog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})
}
