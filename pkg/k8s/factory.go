// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"context"

	vv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	openfaasv1 "github.com/openfaas/faas-netes/pkg/client/clientset/versioned/typed/openfaas/v1"

	v1 "github.com/openfaas/faas-netes/pkg/client/listers/openfaas/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

func NewFunctionFactory(clientset kubernetes.Interface, config DeploymentConfig, faasclient openfaasv1.OpenfaasV1Interface) FunctionFactory {
	return FunctionFactory{
		Client:   clientset,
		Config:   config,
		Profiler: &Lister{f: faasclient},
	}
}

type Lister struct {
	f openfaasv1.OpenfaasV1Interface
}

func (l *Lister) Profiles(namespace string) v1.ProfileNamespaceLister {
	return &NamespaceLister{f: l.f, ns: namespace}
}

type NamespaceLister struct {
	f  openfaasv1.OpenfaasV1Interface
	ns string
}

func (l *NamespaceLister) Get(name string) (ret *vv1.Profile, err error) {
	value, err := l.f.Profiles(l.ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (l *NamespaceLister) List(selector labels.Selector) (ret []*vv1.Profile, err error) {
	list, err := l.f.Profiles(l.ns).List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})

	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		ret = append(ret, &item)
	}

	return ret, nil
}
