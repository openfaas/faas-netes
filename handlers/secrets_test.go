package handlers

import (
	"fmt"
	"testing"

	"github.com/openfaas/faas/gateway/requests"
	apiv1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
)

func Test_UpdateSecretsAddOrUpdatesTheVolumeSpec(t *testing.T) {
	t.Run("No volume added if request secrets is nil", func(t *testing.T) {
		request := requests.CreateFunctionRequest{
			Service: "testfunc",
			Secrets: nil,
		}
		existingSecrets := map[string]*apiv1.Secret{
			"pullsecret": {Type: apiv1.SecretTypeDockercfg},
			"testsecret": {Type: apiv1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
		}

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
		err := UpdateSecrets(request, deployment, existingSecrets)
		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
		}

		validateEmptySecretVolumesAndMounts(t, deployment)

	})

	t.Run("No volume added if request secrets is the empty slice", func(t *testing.T) {
		request := requests.CreateFunctionRequest{
			Service: "testfunc",
			Secrets: []string{},
		}
		existingSecrets := map[string]*apiv1.Secret{
			"pullsecret": {Type: apiv1.SecretTypeDockercfg},
			"testsecret": {Type: apiv1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
		}

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
		err := UpdateSecrets(request, deployment, existingSecrets)
		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
		}

		validateEmptySecretVolumesAndMounts(t, deployment)

	})

	t.Run("Adds New Secrets Volume", func(t *testing.T) {
		request := requests.CreateFunctionRequest{
			Service: "testfunc",
			Secrets: []string{"pullsecret", "testsecret"},
		}
		existingSecrets := map[string]*apiv1.Secret{
			"pullsecret": {Type: apiv1.SecretTypeDockercfg},
			"testsecret": {Type: apiv1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
		}

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
		err := UpdateSecrets(request, deployment, existingSecrets)
		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
		}

		validateNewSecretVolumesAndMounts(t, deployment)

	})

	t.Run("Replaces Previous SecretMount with new mount", func(t *testing.T) {
		request := requests.CreateFunctionRequest{
			Service: "testfunc",
			Secrets: []string{"pullsecret", "testsecret"},
		}
		existingSecrets := map[string]*apiv1.Secret{
			"pullsecret": {Type: apiv1.SecretTypeDockercfg},
			"testsecret": {Type: apiv1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
		}

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
		err := UpdateSecrets(request, deployment, existingSecrets)
		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
		}
		// mimic the deployment already existing and deployed with the same secrets by running
		// UpdateSecrets twice, the first run represents the original deployment, the second run represents
		// retrieving the deployment from the k8s api and applying the update to it
		err = UpdateSecrets(request, deployment, existingSecrets)
		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
		}

		validateNewSecretVolumesAndMounts(t, deployment)

	})

	t.Run("Removes secrets volumes if the request is empty or nil", func(t *testing.T) {
		request := requests.CreateFunctionRequest{
			Service: "testfunc",
			Secrets: []string{"pullsecret", "testsecret"},
		}
		existingSecrets := map[string]*apiv1.Secret{
			"pullsecret": {Type: apiv1.SecretTypeDockercfg},
			"testsecret": {Type: apiv1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
		}

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
		err := UpdateSecrets(request, deployment, existingSecrets)
		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
		}

		validateNewSecretVolumesAndMounts(t, deployment)

		request = requests.CreateFunctionRequest{
			Service: "testfunc",
			Secrets: []string{},
		}
		err = UpdateSecrets(request, deployment, existingSecrets)
		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
		}

		validateEmptySecretVolumesAndMounts(t, deployment)
	})
}

func validateEmptySecretVolumesAndMounts(t *testing.T, deployment *v1beta1.Deployment) {
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

func validateNewSecretVolumesAndMounts(t *testing.T, deployment *v1beta1.Deployment) {
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
}
