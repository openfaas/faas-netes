// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	types "github.com/openfaas/faas-provider/types"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"testing"
)

func readOnlyRootDisabled(t *testing.T, deployment *appsv1.Deployment) {
	if len(deployment.Spec.Template.Spec.Volumes) != 0 {
		t.Error("Volumes should be empty if ReadOnlyRootFilesystem is false")
	}

	if len(deployment.Spec.Template.Spec.Containers[0].VolumeMounts) != 0 {
		t.Error("VolumeMounts should be empty if ReadOnlyRootFilesystem is false")
	}
	functionContatiner := deployment.Spec.Template.Spec.Containers[0]

	if functionContatiner.SecurityContext != nil {
		if *functionContatiner.SecurityContext.ReadOnlyRootFilesystem != false {
			t.Error("ReadOnlyRootFilesystem should be false on the container SecurityContext")
		}
	}
}

func readOnlyRootEnabled(t *testing.T, deployment *appsv1.Deployment) {
	if len(deployment.Spec.Template.Spec.Volumes) != 1 {
		t.Error("should create a single tmp Volume")
	}

	if len(deployment.Spec.Template.Spec.Containers[0].VolumeMounts) != 1 {
		t.Error("should create a single tmp VolumeMount")
	}

	volume := deployment.Spec.Template.Spec.Volumes[0]
	if volume.Name != "temp" {
		t.Error("volume should be named temp")
	}

	mount := deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0]
	if mount.Name != "temp" {
		t.Error("volume mount should be named temp")
	}

	if mount.MountPath != "/tmp" {
		t.Error("temp volume should be mounted to /tmp")
	}

	if mount.ReadOnly {
		t.Errorf("temp mount should not read only")
	}

	if deployment.Spec.Template.Spec.Containers[0].SecurityContext == nil {
		t.Error("container security context should not be nil")
	}

	if *deployment.Spec.Template.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem != true {
		t.Error("should set ReadOnlyRootFilesystem to true on the container SecurityContext")
	}
}

func Test_configureReadOnlyRootFilesystem_Disabled_To_Disabled(t *testing.T) {
	f := mockFactory()
	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}

	request := types.FunctionDeployment{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: false,
	}

	f.ConfigureReadOnlyRootFilesystem(request, deployment)
	readOnlyRootDisabled(t, deployment)
}

func Test_configureReadOnlyRootFilesystem_Disabled_To_Enabled(t *testing.T) {
	f := mockFactory()
	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}

	request := types.FunctionDeployment{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: true,
	}

	f.ConfigureReadOnlyRootFilesystem(request, deployment)
	readOnlyRootEnabled(t, deployment)
}

func Test_configureReadOnlyRootFilesystem_Enabled_To_Disabled(t *testing.T) {
	f := mockFactory()
	trueValue := true
	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "testfunc",
							Image: "alpine:latest",
							SecurityContext: &apiv1.SecurityContext{
								ReadOnlyRootFilesystem: &trueValue,
							},
							VolumeMounts: []apiv1.VolumeMount{
								{Name: "temp", MountPath: "/tmp", ReadOnly: false},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "temp",
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	request := types.FunctionDeployment{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: false,
	}
	f.ConfigureReadOnlyRootFilesystem(request, deployment)
	readOnlyRootDisabled(t, deployment)
}

func Test_configureReadOnlyRootFilesystem_Enabled_To_Enabled(t *testing.T) {
	f := mockFactory()
	trueValue := true
	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "testfunc",
							Image: "alpine:latest",
							SecurityContext: &apiv1.SecurityContext{
								ReadOnlyRootFilesystem: &trueValue,
							},
							VolumeMounts: []apiv1.VolumeMount{
								{Name: "temp", MountPath: "/tmp", ReadOnly: false},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "temp",
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	request := types.FunctionDeployment{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: true,
	}
	f.ConfigureReadOnlyRootFilesystem(request, deployment)
	readOnlyRootEnabled(t, deployment)
}
