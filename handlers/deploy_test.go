package handlers

import (
	"testing"

	"github.com/openfaas/faas/gateway/requests"
	apiv1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
)

func Test_configureReadOnlyRootFilesystem_Disabled_To_Disabled(t *testing.T) {
	deployment := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}

	request := requests.CreateFunctionRequest{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: false,
	}

	configureReadOnlyRootFilesystem(request, deployment)
	readOnlyRootDisabled(t, deployment)
}

func Test_configureReadOnlyRootFilesystem_Disabled_To_Enabled(t *testing.T) {
	deployment := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}

	request := requests.CreateFunctionRequest{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: true,
	}

	configureReadOnlyRootFilesystem(request, deployment)
	readOnlyRootEnabled(t, deployment)
}

func Test_configureReadOnlyRootFilesystem_Enabled_To_Disabled(t *testing.T) {
	trueValue := true
	deployment := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
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

	request := requests.CreateFunctionRequest{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: false,
	}
	configureReadOnlyRootFilesystem(request, deployment)
	readOnlyRootDisabled(t, deployment)
}

func Test_configureReadOnlyRootFilesystem_Enabled_To_Enabled(t *testing.T) {
	trueValue := true
	deployment := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
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

	request := requests.CreateFunctionRequest{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: true,
	}
	configureReadOnlyRootFilesystem(request, deployment)
	readOnlyRootEnabled(t, deployment)
}

func Test_buildAnnotations_Empty_In_CreateRequest(t *testing.T) {
	request := requests.CreateFunctionRequest{}

	annotations := buildAnnotations(request)

	if len(annotations) != 1 {
		t.Errorf("want: %d annotations got: %d", 1, len(annotations))
	}

	v, ok := annotations["prometheus.io.scrape"]
	if !ok {
		t.Errorf("missing prometheus.io.scrape key")
	}

	if v != "false" {
		t.Errorf("want: %s for annotation prometheus.io.scrape got: %s", "false", v)
	}
}

func Test_buildAnnotations_From_CreateRequest(t *testing.T) {
	request := requests.CreateFunctionRequest{
		Annotations: &map[string]string{
			"date-created": "Wed 25 Jul 21:26:22 BST 2018",
			"foo" : "bar",
		},
	}

	annotations := buildAnnotations(request)

	if len(annotations) != 3 {
		t.Errorf("want: %d annotations got: %d", 1, len(annotations))
	}

	v, ok := annotations["date-created"]
	if !ok {
		t.Errorf("missing date-created key")
	}

	if v != "Wed 25 Jul 21:26:22 BST 2018" {
		t.Errorf("want: %s for annotation date-created got: %s", "Wed 25 Jul 21:26:22 BST 2018", v)
	}
}

func readOnlyRootDisabled(t *testing.T, deployment *v1beta1.Deployment) {
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

func readOnlyRootEnabled(t *testing.T, deployment *v1beta1.Deployment) {
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
