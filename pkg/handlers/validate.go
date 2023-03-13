// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright 2020 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"regexp"
	"strconv"

	types "github.com/openfaas/faas-provider/types"
)

// Regex for RFC-1123 validation:
//
//	k8s.io/kubernetes/pkg/util/validation/validation.go
var validDNS = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// validates that the service name is valid for Kubernetes
func validateService(service string) error {
	matched := validDNS.MatchString(service)
	if matched {
		return nil
	}

	return fmt.Errorf("service: (%s) is invalid, must be a valid DNS entry", service)
}

// ValidateDeployRequest validates that the service name is valid for Kubernetes
func ValidateDeployRequest(request *types.FunctionDeployment) error {

	if request.Service == "" {
		return fmt.Errorf("service: is required")
	}

	err := validateService(request.Service)
	if err != nil {
		return err
	}

	if request.Image == "" {
		return fmt.Errorf("image: is required")
	}

	if err := validateScalingLabels(request); err != nil {
		return err
	}

	return nil
}

func validateScalingLabels(request *types.FunctionDeployment) error {
	if request.Labels == nil {
		return nil
	}

	labels := *request.Labels
	if v, ok := labels["com.openfaas.scale.zero"]; ok {
		if v == "true" {
			return fmt.Errorf("com.openfaas.scale.zero not available for Community Edition")
		}
	}
	if _, ok := labels["com.openfaas.scale.zero-duration"]; ok {
		return fmt.Errorf("com.openfaas.scale.zero-duration not available for Community Edition")
	}

	if _, ok := labels["com.openfaas.scale.target"]; ok {
		return fmt.Errorf("com.openfaas.scale.target not available for Community Edition")
	}

	if _, ok := labels["com.openfaas.scale.type"]; ok {
		return fmt.Errorf("com.openfaas.scale.type not available for Community Edition")
	}

	if v, ok := labels["com.openfaas.scale.max"]; ok {
		if vv, err := strconv.Atoi(v); err == nil {
			if vv > MaxReplicas {
				return fmt.Errorf("com.openfaas.scale.max is set too high for Community Edition")
			}
		}
	}

	if v, ok := labels["com.openfaas.scale.min"]; ok {
		if vv, err := strconv.Atoi(v); err == nil {
			if vv > MaxReplicas {
				return fmt.Errorf("com.openfaas.scale.min is set too high for Community Edition")
			}
		}
	}

	return nil
}
