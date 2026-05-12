// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Author(s) 2020. All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	informers "github.com/openfaas/faas-netes/pkg/client/informers/externalversions"
	"github.com/openfaas/faas-netes/pkg/config"
	"github.com/openfaas/faas-netes/pkg/handlers"
	"github.com/openfaas/faas-netes/pkg/k8s"
	"github.com/openfaas/faas-netes/pkg/signals"
	version "github.com/openfaas/faas-netes/version"
	faasProvider "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"
	providertypes "github.com/openfaas/faas-provider/types"

	kubeinformers "k8s.io/client-go/informers"
	v1apps "k8s.io/client-go/informers/apps/v1"
	v1core "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	// required to authenticate against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// main.go:36:2: import "sigs.k8s.io/controller-tools/cmd/controller-gen" is a program, not an importable package
	// _ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)

const defaultResync = time.Hour * 10

func main() {
	var kubeconfig string
	var masterURL string
	var (
		verbose bool
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig. Only required if out-of-cluster.")
	flag.BoolVar(&verbose, "verbose", false, "Print verbose config information")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.Bool("operator", false, "Run as an operator (not available in CE)")

	flag.Parse()

	sha, release := version.GetReleaseInfo()
	fmt.Printf(`faas-netes - Community Edition (CE)
Warning: Commercial use limited to 60 days.
Learn more: https://github.com/openfaas/faas/blob/master/EULA.md

Version: %s Commit: %s
`, release, sha)

	if err := config.ConnectivityCheck(); err != nil {
		log.Fatalf("Error checking connectivity, OpenFaaS CE cannot be run in an offline environment: %s", err.Error())
	}

	clientCmdConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeconfigQPS := 100
	kubeconfigBurst := 250

	clientCmdConfig.QPS = float32(kubeconfigQPS)
	clientCmdConfig.Burst = kubeconfigBurst

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

	// use kubeclient to check the current namespace
	namespace, _ := k8s.CurrentNamespace()
	if namespace == "kube-system" {
		log.Fatal("You cannot run the OpenFaaS provider in the kube-system namespace, please try another namespace.")
	}

	deployConfig := k8s.DeploymentConfig{
		RuntimeHTTPPort: 8080,
		HTTPProbe:       config.HTTPProbe,
		SetNonRootUser:  config.SetNonRootUser,
		ReadinessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(2),
			TimeoutSeconds:      int32(1),
			PeriodSeconds:       int32(2),
		},
		LivenessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(2),
			TimeoutSeconds:      int32(1),
			PeriodSeconds:       int32(2),
		},
	}

	namespaceScope := config.DefaultFunctionNamespace

	if namespaceScope == "" {
		klog.Fatal("DefaultFunctionNamespace must be set")
	}

	kubeInformerOpt := kubeinformers.WithNamespace(namespaceScope)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResync, kubeInformerOpt)

	faasInformerOpt := informers.WithNamespace(namespaceScope)
	faasInformerFactory := informers.NewSharedInformerFactoryWithOptions(faasClient, defaultResync, faasInformerOpt)

	factory := k8s.NewFunctionFactory(kubeClient, deployConfig, faasClient.OpenfaasV1())

	setup := serverSetup{
		config:              config,
		functionFactory:     factory,
		kubeInformerFactory: kubeInformerFactory,
		faasInformerFactory: faasInformerFactory,
		kubeClient:          kubeClient,
		faasClient:          faasClient,
	}

	runController(setup)
}

type customInformers struct {
	EndpointsInformer  v1core.EndpointsInformer
	DeploymentInformer v1apps.DeploymentInformer
}

func startInformers(setup serverSetup, stopCh <-chan struct{}, operator bool) customInformers {
	kubeInformerFactory := setup.kubeInformerFactory

	deployments := kubeInformerFactory.Apps().V1().Deployments()
	go deployments.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync("faas-netes:deployments", stopCh, deployments.Informer().HasSynced); !ok {
		log.Fatalf("failed to wait for cache to sync")
	}

	endpoints := kubeInformerFactory.Core().V1().Endpoints()
	go endpoints.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync("faas-netes:endpoints", stopCh, endpoints.Informer().HasSynced); !ok {
		log.Fatalf("failed to wait for cache to sync")
	}

	return customInformers{
		EndpointsInformer:  endpoints,
		DeploymentInformer: deployments,
	}
}

// runController runs the faas-netes imperative controller
func runController(setup serverSetup) {
	config := setup.config
	kubeClient := setup.kubeClient
	factory := setup.functionFactory

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	operator := false
	listers := startInformers(setup, stopCh, operator)
	handlers.RegisterEventHandlers(listers.DeploymentInformer, kubeClient, config.DefaultFunctionNamespace)
	deployLister := listers.DeploymentInformer.Lister()
	functionLookup := k8s.NewFunctionLookup(config.DefaultFunctionNamespace, listers.EndpointsInformer.Lister())
	functionList := k8s.NewFunctionList(config.DefaultFunctionNamespace, deployLister)

	printFunctionExecutionTime := true

	proxyHandler := proxy.NewHandlerFunc(config.FaaSConfig, functionLookup, printFunctionExecutionTime)

	if err := handlers.Check(functionList); err != nil {
		msg := fmt.Sprintf("Function invocations disabled due to error: %s.", err.Error())
		log.Print(msg)

		proxyHandler = func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, msg, http.StatusMethodNotAllowed)
		}
	}

	bootstrapHandlers := providertypes.FaaSHandlers{
		FunctionProxy:  proxyHandler,
		DeleteFunction: handlers.MakeDeleteHandler(config.DefaultFunctionNamespace, kubeClient),
		DeployFunction: handlers.MakeDeployHandler(config.DefaultFunctionNamespace, factory, functionList),
		FunctionLister: handlers.MakeFunctionReader(config.DefaultFunctionNamespace, deployLister),
		FunctionStatus: handlers.MakeReplicaReader(config.DefaultFunctionNamespace, deployLister),
		ScaleFunction:  handlers.MakeReplicaUpdater(config.DefaultFunctionNamespace, kubeClient),
		UpdateFunction: handlers.MakeUpdateHandler(config.DefaultFunctionNamespace, factory),
		Health:         handlers.MakeHealthHandler(),
		Info:           handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommit),
		Secrets:        handlers.MakeSecretHandler(config.DefaultFunctionNamespace, kubeClient),
		Logs:           logs.NewLogHandlerFunc(k8s.NewLogRequestor(kubeClient, config.DefaultFunctionNamespace), config.FaaSConfig.WriteTimeout),
		ListNamespaces: handlers.MakeNamespacesLister(config.DefaultFunctionNamespace, kubeClient),
	}

	ctx := context.Background()

	faasProvider.Serve(ctx, &bootstrapHandlers, &config.FaaSConfig)
}

// serverSetup is a container for the config and clients needed to start the
// faas-netes controller or operator
type serverSetup struct {
	config              config.BootstrapConfig
	kubeClient          *kubernetes.Clientset
	faasClient          *clientset.Clientset
	functionFactory     k8s.FunctionFactory
	kubeInformerFactory kubeinformers.SharedInformerFactory
	faasInformerFactory informers.SharedInformerFactory
}
