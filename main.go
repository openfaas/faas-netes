// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/openfaas/faas-provider/proxy"
	"k8s.io/client-go/kubernetes"

	"github.com/openfaas-incubator/openfaas-operator/pkg/signals"
	"github.com/openfaas/faas-netes/handlers"
	"github.com/openfaas/faas-netes/k8s"
	"github.com/openfaas/faas-netes/types"
	"github.com/openfaas/faas-netes/version"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	bootTypes "github.com/openfaas/faas-provider/types"
	kubeinformers "k8s.io/client-go/informers"

	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig string
	var masterURL string

	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.Parse()

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
	osEnv := bootTypes.OsEnv{}
	cfg, err := readConfig.Read(osEnv)
	if err != nil {
		log.Fatalf("Error reading config: %s", err.Error())
	}

	log.Printf("HTTP Read Timeout: %s\n", cfg.FaaSConfig.GetReadTimeout())
	log.Printf("HTTP Write Timeout: %s\n", cfg.FaaSConfig.WriteTimeout)
	log.Printf("HTTPProbe: %v\n", cfg.HTTPProbe)
	log.Printf("SetNonRootUser: %v\n", cfg.SetNonRootUser)

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

	bootstrapHandlers := bootTypes.FaaSHandlers{
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

	bootstrap.Serve(&bootstrapHandlers, &cfg.FaaSConfig)
}
