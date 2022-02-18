// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"path/filepath"

	types "github.com/openfaas/faas-provider/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ProbePathValue = "/_/health"
)

type FunctionProbes struct {
	Liveness  *corev1.Probe
	Readiness *corev1.Probe
}

// MakeProbes returns the liveness and readiness probes
// by default the health check runs `cat /tmp/.lock` every ten seconds
func (f *FunctionFactory) MakeProbes(r types.FunctionDeployment) (*FunctionProbes, error) {
	var handler corev1.Handler
	httpPath := ProbePathValue
	initialDelaySeconds := int32(f.Config.LivenessProbe.InitialDelaySeconds)

	if f.Config.HTTPProbe || httpPath != ProbePathValue {
		handler = corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: httpPath,
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: int32(f.Config.RuntimeHTTPPort),
				},
			},
		}
	} else {
		path := filepath.Join("/tmp/", ".lock")
		handler = corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"cat", path},
			},
		}
	}

	probes := FunctionProbes{}
	probes.Readiness = &corev1.Probe{
		Handler:             handler,
		InitialDelaySeconds: initialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.ReadinessProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.ReadinessProbe.PeriodSeconds),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	probes.Liveness = &corev1.Probe{
		Handler:             handler,
		InitialDelaySeconds: initialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.LivenessProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.LivenessProbe.PeriodSeconds),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	return &probes, nil
}
