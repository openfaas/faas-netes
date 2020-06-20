// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	openfaasv1 "github.com/openfaas/faas-netes/pkg/client/clientset/versioned/typed/openfaas/v1"
	"k8s.io/client-go/kubernetes"
)

// FunctionFactory is handling Kubernetes operations to materialise functions into deployments and services
type FunctionFactory struct {
	Client   kubernetes.Interface
	Config   DeploymentConfig
	OFClient openfaasv1.OpenfaasV1Interface
}

func NewFunctionFactory(clientset kubernetes.Interface, config DeploymentConfig, ofClient openfaasv1.OpenfaasV1Interface) FunctionFactory {
	return FunctionFactory{
		Client:   clientset,
		Config:   config,
		OFClient: ofClient,
	}
}
