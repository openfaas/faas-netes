// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Author(s) 2020. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"flag"
	"log"
	"time"

	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	informers "github.com/openfaas/faas-netes/pkg/client/informers/externalversions"
	"github.com/openfaas/faas-netes/pkg/config"
	"github.com/openfaas/faas-netes/pkg/controller"
	"github.com/openfaas/faas-netes/pkg/handlers"
	"github.com/openfaas/faas-netes/pkg/k8s"
	"github.com/openfaas/faas-netes/pkg/server"
	"github.com/openfaas/faas-netes/pkg/signals"
	version "github.com/openfaas/faas-netes/version"
	faasProvider "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"
	providertypes "github.com/openfaas/faas-provider/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	glog "k8s.io/klog"

	// required to authenticate against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// required for updating and validating the CRD clientset
	_ "k8s.io/code-generator/cmd/client-gen/generators"
	// main.go:36:2: import "sigs.k8s.io/controller-tools/cmd/controller-gen" is a program, not an importable package
	// _ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)

func main() {
	var kubeconfig string
	var masterURL string
	var (
		operator,
		verbose bool
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig. Only required if out-of-cluster.")
	flag.BoolVar(&verbose, "verbose", false, "Print verbose config information")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	flag.BoolVar(&operator, "operator", false, "Use the operator mode instead of faas-netes")
	flag.Parse()

	sha, release := version.GetReleaseInfo()
	log.Printf("Version: %s\tcommit: %s\n", release, sha)

	clientCmdConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(clientCmdConfig)
	if err != nil {
		log.Fatalf("Error building Kubernetes clientset: %s", err.Error())
	}

	faasClient, err := clientset.NewForConfig(clientCmdConfig)
	if err != nil {
		log.Fatalf("Error building OpenFaaS clientset: %s", err.Error())
	}

	readConfig := config.ReadConfig{}
	osEnv := providertypes.OsEnv{}
	config, err := readConfig.Read(osEnv)

	if err != nil {
		log.Fatalf("Error reading config: %s", err.Error())
	}

	config.Fprint(verbose)

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
		ImagePullPolicy:   config.ImagePullPolicy,
		ProfilesNamespace: config.ProfilesNamespace,
	}

	// the sync interval does not affect the scale to/from zero feature
	// auto-scaling is does via the HTTP API that acts on the deployment Spec.Replicas
	defaultResync := time.Minute * 5

	namespaceScope := config.DefaultFunctionNamespace
	if config.ClusterRole {
		namespaceScope = ""
	}

	kubeInformerOpt := kubeinformers.WithNamespace(namespaceScope)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResync, kubeInformerOpt)

	faasInformerOpt := informers.WithNamespace(namespaceScope)
	faasInformerFactory := informers.NewSharedInformerFactoryWithOptions(faasClient, defaultResync, faasInformerOpt)

	// this is where we need to swap to the faasInformerFactory
	profileInformerOpt := informers.WithNamespace(config.ProfilesNamespace)
	profileInformerFactory := informers.NewSharedInformerFactoryWithOptions(faasClient, defaultResync, profileInformerOpt)
	profileLister := profileInformerFactory.Openfaas().V1().Profiles().Lister()
	factory := k8s.NewFunctionFactory(kubeClient, deployConfig, profileLister)

	setup := serverSetup{
		config:                 config,
		functionFactory:        factory,
		kubeInformerFactory:    kubeInformerFactory,
		faasInformerFactory:    faasInformerFactory,
		profileInformerFactory: profileInformerFactory,
		kubeClient:             kubeClient,
		faasClient:             faasClient,
	}

	if !operator {
		log.Println("Starting controller")
		runController(setup)
	} else {
		log.Println("Starting operator")
		runOperator(setup, config)
	}
}

// runController runs the faas-netes imperative controller
func runController(setup serverSetup) {
	// pull out the required config and clients fromthe setup, this is largely a
	// leftover from refactoring the setup to a shared step and keeping the function
	// signature readable
	config := setup.config
	kubeClient := setup.kubeClient
	factory := setup.functionFactory
	kubeInformerFactory := setup.kubeInformerFactory
	faasInformerFactory := setup.faasInformerFactory

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	endpointsInformer := kubeInformerFactory.Core().V1().Endpoints()

	deploymentLister := kubeInformerFactory.Apps().V1().
		Deployments().Lister()

	go faasInformerFactory.Start(stopCh)
	go kubeInformerFactory.Start(stopCh)
	go setup.profileInformerFactory.Start(stopCh)

	// Any "Wait" calls need to be made, after the informers have been started
	start := time.Now()
	glog.Infof("Waiting for cache sync in main")
	kubeInformerFactory.WaitForCacheSync(stopCh)
	setup.profileInformerFactory.WaitForCacheSync(stopCh)

	// Block and wait for the endpoints to become synchronised
	cache.WaitForCacheSync(stopCh, endpointsInformer.Informer().HasSynced)

	glog.Infof("Cache sync done in: %fs", time.Since(start).Seconds())
	glog.Infof("Endpoints synced? %v", endpointsInformer.Informer().HasSynced())

	lister := endpointsInformer.Lister()
	functionLookup := k8s.NewFunctionLookup(config.DefaultFunctionNamespace, lister)

	bootstrapHandlers := providertypes.FaaSHandlers{
		FunctionProxy:        proxy.NewHandlerFunc(config.FaaSConfig, functionLookup),
		DeleteHandler:        handlers.MakeDeleteHandler(config.DefaultFunctionNamespace, kubeClient),
		DeployHandler:        handlers.MakeDeployHandler(config.DefaultFunctionNamespace, factory),
		FunctionReader:       handlers.MakeFunctionReader(config.DefaultFunctionNamespace, deploymentLister),
		ReplicaReader:        handlers.MakeReplicaReader(config.DefaultFunctionNamespace, deploymentLister),
		ReplicaUpdater:       handlers.MakeReplicaUpdater(config.DefaultFunctionNamespace, kubeClient),
		UpdateHandler:        handlers.MakeUpdateHandler(config.DefaultFunctionNamespace, factory),
		HealthHandler:        handlers.MakeHealthHandler(),
		InfoHandler:          handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommit),
		SecretHandler:        handlers.MakeSecretHandler(config.DefaultFunctionNamespace, kubeClient),
		LogHandler:           logs.NewLogHandlerFunc(k8s.NewLogRequestor(kubeClient, config.DefaultFunctionNamespace), config.FaaSConfig.WriteTimeout),
		ListNamespaceHandler: handlers.MakeNamespacesLister(config.DefaultFunctionNamespace, config.ClusterRole, kubeClient),
	}

	faasProvider.Serve(&bootstrapHandlers, &config.FaaSConfig)
}

// runOperator runs the CRD Operator
func runOperator(setup serverSetup, cfg config.BootstrapConfig) {
	// pull out the required config and clients fromthe setup, this is largely a
	// leftover from refactoring the setup to a shared step and keeping the function
	// signature readable
	kubeClient := setup.kubeClient
	faasClient := setup.faasClient
	kubeInformerFactory := setup.kubeInformerFactory
	faasInformerFactory := setup.faasInformerFactory

	// the operator wraps the FunctionFactory with its own type
	factory := controller.FunctionFactory{
		Factory: setup.functionFactory,
	}

	setupLogging()
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	endpointsInformer := kubeInformerFactory.Core().V1().Endpoints()
	deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()

	ctrl := controller.NewController(
		kubeClient,
		faasClient,
		kubeInformerFactory,
		faasInformerFactory,
		factory,
	)

	srv := server.New(faasClient, kubeClient, endpointsInformer, deploymentInformer, cfg.ClusterRole, cfg)

	go faasInformerFactory.Start(stopCh)
	go kubeInformerFactory.Start(stopCh)
	go setup.profileInformerFactory.Start(stopCh)

	// Any "Wait" calls need to be made, after the informers have been started
	start := time.Now()
	glog.Infof("Waiting for cache sync in main")
	kubeInformerFactory.WaitForCacheSync(stopCh)
	setup.profileInformerFactory.WaitForCacheSync(stopCh)

	// Block and wait for the endpoints to become synchronised
	cache.WaitForCacheSync(stopCh, endpointsInformer.Informer().HasSynced)

	glog.Infof("Cache sync done in: %fs", time.Since(start).Seconds())
	glog.Infof("Endpoints synced? %v", endpointsInformer.Informer().HasSynced())

	go srv.Start()
	if err := ctrl.Run(1, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

// serverSetup is a container for the config and clients needed to start the
// faas-netes controller or operator
type serverSetup struct {
	config                 config.BootstrapConfig
	kubeClient             *kubernetes.Clientset
	faasClient             *clientset.Clientset
	functionFactory        k8s.FunctionFactory
	kubeInformerFactory    kubeinformers.SharedInformerFactory
	faasInformerFactory    informers.SharedInformerFactory
	profileInformerFactory informers.SharedInformerFactory
}

func setupLogging() {
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	glog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			_ = f2.Value.Set(value)
		}
	})
}
