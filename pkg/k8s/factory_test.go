// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import "k8s.io/client-go/kubernetes/fake"

func mockFactory() FunctionFactory {
	return NewFunctionFactory(fake.NewSimpleClientset(),
		DeploymentConfig{
			HTTPProbe: false,
			LivenessProbe: &ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
			ReadinessProbe: &ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
		}, nil)
}
