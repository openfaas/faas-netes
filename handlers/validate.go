// Copyright (c) Alex Ellis 2017. All rights reserved.
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

// ValidateDeployRequest validates that the service name is valid for Kubernetes
func ValidateDeployRequest(request *types.FunctionDeployment) error {
	matched := validDNS.MatchString(request.Service)
	if matched {
		return nil
	}

	return fmt.Errorf("(%s) must be a valid DNS entry for service name", request.Service)
}
