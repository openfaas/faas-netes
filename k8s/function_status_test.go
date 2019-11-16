// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Env: []corev1.EnvVar{
								{Name: "fprocess", Value: envProcess},
								{Name: "customEnv", Value: "customValue"},
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

}
