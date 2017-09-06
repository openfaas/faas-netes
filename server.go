// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"time"

	"github.com/alexellis/faas-netes/handlers"
	"github.com/alexellis/faas-provider"
	bootTypes "github.com/alexellis/faas-provider/types"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	bootstrapHandlers := bootTypes.FaaSHandlers{
		FunctionProxy:  handlers.MakeProxy(),
		DeleteHandler:  handlers.MakeDeleteHandler(clientset),
		DeployHandler:  handlers.MakeDeployHandler(clientset),
		FunctionReader: handlers.MakeFunctionReader(clientset),
		ReplicaReader:  handlers.MakeReplicaReader(clientset),
		ReplicaUpdater: handlers.MakeReplicaUpdater(clientset),
	}
	var port int
	port = 8080
	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:  time.Second * 8,
		WriteTimeout: time.Second * 8,
		TCPPort:      &port,
	}

	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)

}
