// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"k8s.io/client-go/kubernetes"
)

// FunctionFactory is handling Kubernetes operations to materialise functions into deployments and services
type FunctionFactory struct {
	Client kubernetes.Interface
	Config DeploymentConfig
}

func NewFunctionFactory(clientset kubernetes.Interface, config DeploymentConfig) FunctionFactory {
	return FunctionFactory{
		Client: clientset,
		Config: config,
	}
}
