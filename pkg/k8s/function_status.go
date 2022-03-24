// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	types "github.com/openfaas/faas-provider/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// EnvProcessName is the name of the env variable containing the function process
const EnvProcessName = "fprocess"

// AsFunctionStatus reads a Deployment object into an OpenFaaS FunctionStatus, parsing the
// Deployment and Container spec into a simplified summary of the Function
func AsFunctionStatus(item appsv1.Deployment) *types.FunctionStatus {
	var replicas uint64
	if item.Spec.Replicas != nil {
		replicas = uint64(*item.Spec.Replicas)
	}

	_, functionContainer := FunctionContainer(item)

	labels := item.Spec.Template.Labels
	function := types.FunctionStatus{
		Name:                   item.Name,
		Replicas:               replicas,
		Image:                  functionContainer.Image,
		AvailableReplicas:      uint64(item.Status.AvailableReplicas),
		InvocationCount:        0,
		Labels:                 &labels,
		Annotations:            &item.Spec.Template.Annotations,
		Namespace:              item.Namespace,
		Secrets:                ReadFunctionSecretsSpec(item),
		CreatedAt:              item.CreationTimestamp.Time,
		Constraints:            nodeSelectorToConstraint(item),
		ReadOnlyRootFilesystem: hasReadOnlyRootFilesystem(item),
		EnvVars:                map[string]string{},
	}

	req := &types.FunctionResources{Memory: functionContainer.Resources.Requests.Memory().String(), CPU: functionContainer.Resources.Requests.Cpu().String()}
	lim := &types.FunctionResources{Memory: functionContainer.Resources.Limits.Memory().String(), CPU: functionContainer.Resources.Limits.Cpu().String()}

	if req.CPU != "0" || req.Memory != "0" {
		function.Requests = req
	}
	if lim.CPU != "0" || lim.Memory != "0" {
		function.Limits = lim
	}

	for _, v := range functionContainer.Env {
		if EnvProcessName == v.Name {
			function.EnvProcess = v.Value
			continue
		}
		function.EnvVars[v.Name] = v.Value
	}

	return &function
}

func nodeSelectorToConstraint(deploy appsv1.Deployment) []string {
	nodeSelector := deploy.Spec.Template.Spec.NodeSelector
	constraints := []string{}
	for k, v := range nodeSelector {
		constraints = append(constraints, k+"="+v)
	}
	return constraints
}

func hasReadOnlyRootFilesystem(function appsv1.Deployment) bool {
	_, c := FunctionContainer(function)
	securityContext := c.SecurityContext
	if securityContext == nil {
		return false
	}

	if securityContext.ReadOnlyRootFilesystem == nil {
		return false
	}

	return *securityContext.ReadOnlyRootFilesystem
}

// FunctionContainer returns the index and spec of the OpenFaaS function container
// in the deployment. Use this method to safely retrieve the container spec, it protects
// the controller from potential changes in the deployment spec by other Operators and
// Admission webhooks in the cluster.
//
// idx will be -1 if the function container is not found.
func FunctionContainer(function appsv1.Deployment) (idx int, c corev1.Container) {
	for idx, container := range function.Spec.Template.Spec.Containers {
		if container.Name == function.Name {
			return idx, container
		}
	}
	return -1, c
}
