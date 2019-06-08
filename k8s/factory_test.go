// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

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
		})
}
