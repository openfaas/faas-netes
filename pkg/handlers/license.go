// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package handlers

import (
	"fmt"

	"github.com/openfaas/faas-netes/pkg/k8s"
)

// Check checks the OpenFaaS installation against the OpenFaaS CE EULA
// which this code is licensed under.
func Check(functionList *k8s.FunctionList) error {
	total, err := functionList.Count()
	if err != nil {
		return err
	}

	if total > MaxFunctions {
		return fmt.Errorf("the OpenFaaS CE EULA only permits: %d functions, visit https://openfaas.com/pricing to upgrade to OpenFaaS Standard", MaxFunctions)
	}

	return nil
}
