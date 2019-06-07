package k8s

import "k8s.io/client-go/kubernetes/fake"

func mockFactory() Factory {
	return NewFactory(fake.NewSimpleClientset(),
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
