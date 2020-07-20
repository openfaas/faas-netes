// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	v1 "github.com/openfaas/faas-netes/pkg/client/listers/openfaas/v1"
	"k8s.io/client-go/kubernetes"
)

// NamespacedProfiler is a subset of the v1.ProfileLister that is needed for the function factory
// to support Profiles
type NamespacedProfiler interface {
	Profiles(namespace string) v1.ProfileNamespaceLister
}

// FunctionFactory is handling Kubernetes operations to materialise functions into deployments and services
type FunctionFactory struct {
	Client   kubernetes.Interface
	Config   DeploymentConfig
	Profiler NamespacedProfiler
}

func NewFunctionFactory(clientset kubernetes.Interface, config DeploymentConfig, profiler NamespacedProfiler) FunctionFactory {
	return FunctionFactory{
		Client:   clientset,
		Config:   config,
		Profiler: profiler,
	}
}
