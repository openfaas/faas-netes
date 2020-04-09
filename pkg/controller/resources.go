package controller

import (
	faasv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// makeResources creates deployment resource limits and requests requirements from function specs
func makeResources(function *faasv1.Function) (*corev1.ResourceRequirements, error) {
	resources := &corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	// Set Memory limits
	if function.Spec.Limits != nil && len(function.Spec.Limits.Memory) > 0 {
		qty, err := resource.ParseQuantity(function.Spec.Limits.Memory)
		if err != nil {
			return resources, err
		}
		resources.Limits[corev1.ResourceMemory] = qty
	}
	if function.Spec.Requests != nil && len(function.Spec.Requests.Memory) > 0 {
		qty, err := resource.ParseQuantity(function.Spec.Requests.Memory)
		if err != nil {
			return resources, err
		}
		resources.Requests[corev1.ResourceMemory] = qty
	}

	// Set CPU limits
	if function.Spec.Limits != nil && len(function.Spec.Limits.CPU) > 0 {
		qty, err := resource.ParseQuantity(function.Spec.Limits.CPU)
		if err != nil {
			return resources, err
		}
		resources.Limits[corev1.ResourceCPU] = qty
	}
	if function.Spec.Requests != nil && len(function.Spec.Requests.CPU) > 0 {
		qty, err := resource.ParseQuantity(function.Spec.Requests.CPU)
		if err != nil {
			return resources, err
		}
		resources.Requests[corev1.ResourceCPU] = qty
	}

	return resources, nil
}
