// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright 2020 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"regexp"

	types "github.com/openfaas/faas-provider/types"
)

// Regex for RFC-1123 validation:
// 	k8s.io/kubernetes/pkg/util/validation/validation.go
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

	return nil
}
