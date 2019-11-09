// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"fmt"
	"testing"

	types "github.com/openfaas/faas-provider/types"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

func Test_FunctionFactory_ConfigureSecrets(t *testing.T) {
	f := mockFactory()
	existingSecrets := map[string]*apiv1.Secret{
		"pullsecret": &apiv1.Secret{Type: apiv1.SecretTypeDockercfg},
		"testsecret": &apiv1.Secret{Type: apiv1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
	}

	basicDeployment := appsv1.Deployment{
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

	volumeName := "testfunc-projected-secrets"
	withExistingSecret := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "testfunc",
							Image: "alpine:latest",
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name: volumeName,
								},
								{
									Name: volumeName,
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: volumeName,
						},
						{
							Name: volumeName,
						},
					},
				},
			},
		},
	}

	cases := []struct {
		name       string
		req        types.FunctionDeployment
		deployment appsv1.Deployment
		validator  func(t *testing.T, deployment *appsv1.Deployment)
		err        error
	}{
		{
			name: "does not add volume if request secrets is nil",
			req: types.FunctionDeployment{
				Service: "testfunc",
				Secrets: nil,
			},
			deployment: basicDeployment,
			validator:  validateEmptySecretVolumesAndMounts,
		},
		{
			name: "does not add volume if request secrets is nil",
			req: types.FunctionDeployment{
				Service: "testfunc",
				Secrets: []string{},
			},
			deployment: basicDeployment,
			validator:  validateEmptySecretVolumesAndMounts,
		},
		{
			name: "removes all copies of exiting secrets volumes",
			req: types.FunctionDeployment{
				Service: "testfunc",
				Secrets: []string{},
			},
			deployment: withExistingSecret,
			validator:  validateEmptySecretVolumesAndMounts,
		},
		{
			name: "add new secret volume",
			req: types.FunctionDeployment{
				Service: "testfunc",
				Secrets: []string{"pullsecret", "testsecret"},
			},
			deployment: basicDeployment,
			validator:  validateNewSecretVolumesAndMounts,
		},
		{
			name: "replaces previous secret mount with new mount",
			req: types.FunctionDeployment{
				Service: "testfunc",
				Secrets: []string{"pullsecret", "testsecret"},
			},
			deployment: withExistingSecret,
			validator:  validateNewSecretVolumesAndMounts,
		},
		{
			name: "removes secrets volume if request secrets is empty or nil",
			req: types.FunctionDeployment{
				Service: "testfunc",
				Secrets: []string{},
			},
			deployment: withExistingSecret,
			validator:  validateEmptySecretVolumesAndMounts,
		},
	}

	for _, tc := range cases {
		err := f.ConfigureSecrets(tc.req, &tc.deployment, existingSecrets)
		if err != tc.err {
			t.Errorf("unexpected error result: got %v, expected %v", err, tc.err)
		}

		tc.validator(t, &tc.deployment)
	}
}

func validateEmptySecretVolumesAndMounts(t *testing.T, deployment *appsv1.Deployment) {
	numVolumes := len(deployment.Spec.Template.Spec.Volumes)
	if numVolumes != 0 {
		fmt.Printf("%+v", deployment.Spec.Template.Spec.Volumes)
		t.Errorf("Incorrect number of volumes: expected 0, got %d", numVolumes)
	}

	c := deployment.Spec.Template.Spec.Containers[0]
	numVolumeMounts := len(c.VolumeMounts)
	if numVolumeMounts != 0 {
		t.Errorf("Incorrect number of volumes mounts: expected 0, got %d", numVolumeMounts)
	}
}

func validateNewSecretVolumesAndMounts(t *testing.T, deployment *appsv1.Deployment) {
	numVolumes := len(deployment.Spec.Template.Spec.Volumes)
	if numVolumes != 1 {
		t.Errorf("Incorrect number of volumes: expected 1, got %d", numVolumes)
	}

	volume := deployment.Spec.Template.Spec.Volumes[0]
	if volume.Name != "testfunc-projected-secrets" {
		t.Errorf("Incorrect volume name: expected \"testfunc-projected-secrets\", got \"%s\"", volume.Name)
	}

	if volume.VolumeSource.Projected == nil {
		t.Error("Secrets volume is not a projected volume type")
	}

	if volume.VolumeSource.Projected.Sources[0].Secret.Items[0].Key != "filename" {
		t.Error("Project secret not constructed correctly")
	}

	c := deployment.Spec.Template.Spec.Containers[0]
	numVolumeMounts := len(c.VolumeMounts)
	if numVolumeMounts != 1 {
		t.Errorf("Incorrect number of volumes mounts: expected 1, got %d", numVolumeMounts)
	}

	mount := c.VolumeMounts[0]
	if mount.Name != "testfunc-projected-secrets" {
		t.Errorf("Incorrect volume mounts: expected \"testfunc-projected-secrets\", got \"%s\"", mount.Name)
	}

	if mount.MountPath != secretsMountPath {
		t.Errorf("Incorrect volume mount path: expected \"%s\", got \"%s\"", secretsMountPath, mount.MountPath)
	}
}
