package controller

import (
	"fmt"

	faasv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	secretsMountPath = "/var/openfaas/secrets"
)

// UpdateSecrets will update the Deployment spec to include secrets that have been deployed
// in the kubernetes cluster.  For each requested secret, we inspect the type and add it to the
// deployment spec as appropriate: secrets with type `SecretTypeDockercfg` are added as ImagePullSecrets
// all other secrets are mounted as files in the deployments containers.
func UpdateSecrets(function *faasv1.Function, deployment *appsv1.Deployment, existingSecrets map[string]*corev1.Secret) error {
	// Add / reference pre-existing secrets within Kubernetes
	secretVolumeProjections := []corev1.VolumeProjection{}

	for _, secretName := range function.Spec.Secrets {
		deployedSecret, ok := existingSecrets[secretName]
		if !ok {
			return fmt.Errorf("required secret '%s' was not found in the cluster", secretName)
		}

		switch deployedSecret.Type {

		case corev1.SecretTypeDockercfg,
			corev1.SecretTypeDockerConfigJson:

			deployment.Spec.Template.Spec.ImagePullSecrets = append(
				deployment.Spec.Template.Spec.ImagePullSecrets,
				corev1.LocalObjectReference{
					Name: secretName,
				},
			)

			break

		default:

			projectedPaths := []corev1.KeyToPath{}
			for secretKey := range deployedSecret.Data {
				projectedPaths = append(projectedPaths, corev1.KeyToPath{Key: secretKey, Path: secretKey})
			}

			projection := &corev1.SecretProjection{Items: projectedPaths}
			projection.Name = secretName
			secretProjection := corev1.VolumeProjection{
				Secret: projection,
			}
			secretVolumeProjections = append(secretVolumeProjections, secretProjection)

			break
		}
	}

	volumeName := fmt.Sprintf("%s-projected-secrets", function.Spec.Name)
	projectedSecrets := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
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
	updatedContainers := []corev1.Container{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		mount := corev1.VolumeMount{
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
func removeVolume(volumeName string, volumes []corev1.Volume) []corev1.Volume {
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
func removeVolumeMount(volumeName string, mounts []corev1.VolumeMount) []corev1.VolumeMount {
	newMounts := mounts[:0]
	for _, v := range mounts {
		if v.Name != volumeName {
			newMounts = append(newMounts, v)
		}
	}

	return newMounts
}
