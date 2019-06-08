// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"k8s.io/client-go/kubernetes"
)

// Factory is handling Kubernetes operations to materialise functions into deployments and services
type Factory struct {
	Client kubernetes.Interface
	Config DeploymentConfig
}

func NewFactory(clientset kubernetes.Interface, config DeploymentConfig) Factory {
	return Factory{
		Client: clientset,
		Config: config,
	}
}
