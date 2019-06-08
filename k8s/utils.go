// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

// removeVolume returns a Volume slice with any volumes matching volumeName removed.
// Uses the filter without allocation technique
// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
func removeVolume(volumeName string, volumes []corev1.Volume) []corev1.Volume {
	if volumes == nil {
		return []corev1.Volume{}
	}

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
	if mounts == nil {
		return []corev1.VolumeMount{}
	}

	newMounts := mounts[:0]
	for _, v := range mounts {
		if v.Name != volumeName {
			newMounts = append(newMounts, v)
		}
	}

	return newMounts
}
