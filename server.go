// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"log"
	"os"

	"github.com/openfaas/faas-netes/k8s"

	"github.com/openfaas/faas-netes/handlers"
	"github.com/openfaas/faas-netes/types"
	"github.com/openfaas/faas-netes/version"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	bootTypes "github.com/openfaas/faas-provider/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	functionNamespace := "default"

	if namespace, exists := os.LookupEnv("function_namespace"); exists {
		functionNamespace = namespace
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	readConfig := types.ReadConfig{}
	osEnv := types.OsEnv{}
	cfg := readConfig.Read(osEnv)

	log.Printf("HTTP Read Timeout: %s\n", cfg.ReadTimeout)
	log.Printf("HTTP Write Timeout: %s\n", cfg.WriteTimeout)
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

	bootstrapHandlers := bootTypes.FaaSHandlers{
		FunctionProxy:  handlers.MakeProxy(functionNamespace, cfg.ReadTimeout),
		DeleteHandler:  handlers.MakeDeleteHandler(functionNamespace, clientset),
		DeployHandler:  handlers.MakeDeployHandler(functionNamespace, factory),
		FunctionReader: handlers.MakeFunctionReader(functionNamespace, clientset),
		ReplicaReader:  handlers.MakeReplicaReader(functionNamespace, clientset),
		ReplicaUpdater: handlers.MakeReplicaUpdater(functionNamespace, clientset),
		UpdateHandler:  handlers.MakeUpdateHandler(functionNamespace, factory),
		HealthHandler:  handlers.MakeHealthHandler(),
		InfoHandler:    handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommit),
		SecretHandler:  handlers.MakeSecretHandler(functionNamespace, clientset),
		LogHandler:     logs.NewLogHandlerFunc(handlers.NewLogRequestor(clientset, functionNamespace), cfg.WriteTimeout),
	}

	var port int
	port = cfg.Port

	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		TCPPort:      &port,
		EnableHealth: true,
	}

	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
}
