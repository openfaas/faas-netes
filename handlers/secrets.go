package handlers

import (
	"fmt"

	"github.com/openfaas/faas/gateway/requests"
	apiv1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	secretsMountPath = "/var/openfaas/secrets"
)

// getSecrets queries Kubernetes for a list of secrets by name in the given k8s namespace.
func getSecrets(clientset *kubernetes.Clientset, namespace string, secretNames []string) (map[string]*apiv1.Secret, error) {
	secrets := map[string]*apiv1.Secret{}

	for _, secretName := range secretNames {
		secret, err := clientset.Core().Secrets(namespace).Get(secretName, metav1.GetOptions{})
		if err != nil {
			return secrets, err
		}
		secrets[secretName] = secret
	}

	return secrets, nil
}

// UpdateSecrets will update the Deployment spec to include secrets that have beenb deployed
// in the kubernetes cluster.  For each requested secret, we inspect the type and add it to the
// deployment spec as appropriat: secrets with type `SecretTypeDockercfg/SecretTypeDockerjson`
// are added as ImagePullSecrets all other secrets are mounted as files in the deployments containers.
func UpdateSecrets(request requests.CreateFunctionRequest, deployment *v1beta1.Deployment, existingSecrets map[string]*apiv1.Secret) error {
	// Add / reference pre-existing secrets within Kubernetes
	secretVolumeProjections := []apiv1.VolumeProjection{}

	for _, secretName := range request.Secrets {
		deployedSecret, ok := existingSecrets[secretName]
		if !ok {
			return fmt.Errorf("Required secret '%s' was not found in the cluster", secretName)
		}

		switch deployedSecret.Type {

		case apiv1.SecretTypeDockercfg,
			apiv1.SecretTypeDockerConfigJson:

			deployment.Spec.Template.Spec.ImagePullSecrets = append(
				deployment.Spec.Template.Spec.ImagePullSecrets,
				apiv1.LocalObjectReference{
					Name: secretName,
				},
			)

			break

		default:

			projectedPaths := []apiv1.KeyToPath{}
			for secretKey := range deployedSecret.Data {
				projectedPaths = append(projectedPaths, apiv1.KeyToPath{Key: secretKey, Path: secretKey})
			}

			projection := &apiv1.SecretProjection{Items: projectedPaths}
			projection.Name = secretName
			secretProjection := apiv1.VolumeProjection{
				Secret: projection,
			}
			secretVolumeProjections = append(secretVolumeProjections, secretProjection)

			break
		}
	}

	volumeName := fmt.Sprintf("%s-projected-secrets", request.Service)
	projectedSecrets := apiv1.Volume{
		Name: volumeName,
		VolumeSource: apiv1.VolumeSource{
			Projected: &apiv1.ProjectedVolumeSource{
				Sources: secretVolumeProjections,
			},
		},
	}

	// remove the existing secrets volume, if we can find it. The update volume will be
	// added below
	existingVolumes := removeVolume(volumeName, deployment.Spec.Template.Spec.Volumes)
	deployment.Spec.Template.Spec.Volumes = existingVolumes
	if len(secretVolumeProjections) > 0 {
		deployment.Spec.Template.Spec.Volumes = append(existingVolumes, projectedSecrets)
	}

	// add mount secret as a file
	updatedContainers := []apiv1.Container{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		mount := apiv1.VolumeMount{
			Name:      volumeName,
			ReadOnly:  true,
			MountPath: secretsMountPath,
		}

		// remove the existing secrets volume mount, if we can find it. We update it later.
		container.VolumeMounts = removeVolumeMount(volumeName, container.VolumeMounts)
		if len(secretVolumeProjections) > 0 {
			container.VolumeMounts = append(container.VolumeMounts, mount)
		}

		updatedContainers = append(updatedContainers, container)
	}

	deployment.Spec.Template.Spec.Containers = updatedContainers

	return nil
}

// removeVolume returns a Volume slice with any volumes matching volumeName removed.
// Uses the filter without allocation technique
// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
func removeVolume(volumeName string, volumes []apiv1.Volume) []apiv1.Volume {
	newVolumes := volumes[:0]
	for _, v := range volumes {
		if v.Name != volumeName {
			newVolumes = append(newVolumes, v)
		}
	}

	return newVolumes
}

// removeVolumeMount returns a VolumeMount slice with any mounts matching volumeName removed
// Uses the filter without allocation technique
// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
func removeVolumeMount(volumeName string, mounts []apiv1.VolumeMount) []apiv1.VolumeMount {
	newMounts := mounts[:0]
	for _, v := range mounts {
		if v.Name != volumeName {
			newMounts = append(newMounts, v)
		}
	}

	return newMounts
}
