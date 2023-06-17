// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"fmt"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_AsFunctionStatus(t *testing.T) {

	name := "test-func"
	replicas := int32(2)
	image := "openfaas/func:latest"
	availableReplicas := int32(1)
	labels := map[string]string{"foo": "foovalue"}
	annotations := map[string]string{"data": "datavalue"}
	namespace := "func-namespace"
	envProcess := "process string here"
	envVars := map[string]string{"env1": "env1value", "env2": "env2value"}
	secrets := []string{"0-imagepullsecret", "1-genericsecret", "2-genericsecret"}
	readOnlyRootFilesystem := false
	constraints := []string{"node-label=node-value"}

	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"node-label": "node-value",
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "0-imagepullsecret"},
					},
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Env: []corev1.EnvVar{
								{Name: "fprocess", Value: envProcess},
								{Name: "env1", Value: "env1value"},
								{Name: "env2", Value: "env2value"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: fmt.Sprintf(secretsProjectVolumeNameTmpl, name),
							VolumeSource: corev1.VolumeSource{
								Projected: &corev1.ProjectedVolumeSource{
									Sources: []corev1.VolumeProjection{
										{
											Secret: &corev1.SecretProjection{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: "2-genericsecret",
												},
											},
										},
										{
											Secret: &corev1.SecretProjection{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: "1-genericsecret",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: availableReplicas,
		},
	}
	status := AsFunctionStatus(deploy)

	if status.Name != name {
		t.Errorf("incorrect Name: expected %s, got %s", name, status.Name)
	}

	if status.Replicas != uint64(replicas) {
		t.Errorf("incorrect Replicas: expected %d, got %d", replicas, status.Replicas)
	}

	if status.Image != image {
		t.Errorf("incorrect Image: expected %s, got %s", image, status.Image)
	}

	if status.AvailableReplicas != uint64(availableReplicas) {
		t.Errorf("incorrect AvailableReplicas: expected %d, got %d", availableReplicas, status.AvailableReplicas)
	}

	if !reflect.DeepEqual(status.Labels, &labels) {
		t.Errorf("incorrect Labels: expected %+v, got %+v", labels, status.Labels)
	}

	if !reflect.DeepEqual(status.Annotations, &annotations) {
		t.Errorf("incorrect Annotations: expected %+v, got %+v", annotations, status.Annotations)
	}

	if status.Namespace != namespace {
		t.Errorf("incorrect Namespace: expected %s, got %s", namespace, status.Namespace)
	}

	if status.EnvProcess != envProcess {
		t.Errorf("incorrect EnvProcess: expected %s, got %s", envProcess, status.EnvProcess)
	}

	if !reflect.DeepEqual(status.Secrets, secrets) {
		t.Errorf("incorrect Secrets: expected %v, got : %v", secrets, status.Secrets)
	}

	if !reflect.DeepEqual(status.EnvVars, envVars) {
		t.Errorf("incorrect EnvVars: expected %+v, got %+v", envVars, status.EnvVars)
	}

	if !reflect.DeepEqual(status.Constraints, constraints) {
		t.Errorf("incorrect Constraints: expected %+v, got %+v", constraints, status.Constraints)
	}

	if status.ReadOnlyRootFilesystem != readOnlyRootFilesystem {
		t.Errorf("incorrect ReadOnlyRootFilesystem: expected %v, got : %v", readOnlyRootFilesystem, status.ReadOnlyRootFilesystem)
	}

	t.Run("can parse readonly root filesystem when nil", func(t *testing.T) {
		readOnlyRootFilesystem := false
		deployment := deploy
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}

		status := AsFunctionStatus(deployment)
		if status.ReadOnlyRootFilesystem != readOnlyRootFilesystem {
			t.Errorf("incorrect ReadOnlyRootFilesystem: expected %v, got : %v", readOnlyRootFilesystem, status.ReadOnlyRootFilesystem)
		}
	})

	t.Run("can parse readonly root filesystem enabled", func(t *testing.T) {
		readOnlyRootFilesystem := true
		deployment := deploy
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
		}

		status := AsFunctionStatus(deployment)
		if status.ReadOnlyRootFilesystem != readOnlyRootFilesystem {
			t.Errorf("incorrect ReadOnlyRootFilesystem: expected %v, got : %v", readOnlyRootFilesystem, status.ReadOnlyRootFilesystem)
		}
	})

	t.Run("returns non-empty resource requests", func(t *testing.T) {
		deployment := deploy
		deployment.Spec.Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("100Mi"),
		}

		status := AsFunctionStatus(deployment)
		if status.Requests.CPU != "100m" {
			t.Errorf("incorrect Requests.CPU: expected %s, got %s", "100m", status.Requests.CPU)
		}

		if status.Requests.Memory != "100Mi" {
			t.Errorf("incorrect Requests.Memory: expected %s, got %s", "100Mi", status.Requests.Memory)
		}
	})

	t.Run("returns non-empty resource limits", func(t *testing.T) {
		deployment := deploy
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("100Mi"),
		}

		status := AsFunctionStatus(deployment)
		if status.Limits.CPU != "100m" {
			t.Errorf("incorrect Limits.CPU: expected %s, got %s", "100m", status.Limits.CPU)
		}

		if status.Limits.Memory != "100Mi" {
			t.Errorf("incorrect Limits.Memory: expected %s, got %s", "100Mi", status.Limits.Memory)
		}
	})
}
