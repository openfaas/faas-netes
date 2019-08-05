// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package test

import (
	"testing"

	"github.com/openfaas/faas-netes/handlers"
	types "github.com/openfaas/faas-provider/types"
)

func Test_ValidateDeployRequest_ValidCharacters(t *testing.T) {
	cases := []struct {
		scenario string
		value    string
	}{
		{"lower", "abz"},
		{"includes hyphen", "test-function"},
		{"can start with a digit", "1abz"},
		{"can end with a digit", "abz1"},
	}

	for _, testCase := range cases {
		request := types.FunctionDeployment{
			Service: testCase.value,
		}

		err := handlers.ValidateDeployRequest(&request)
		if err != nil {
			t.Errorf("Scenario: %s with value: %s, got: %s", testCase.scenario, testCase.value, err.Error())
		}

	}
}

func Test_ValidateDeployRequest_InvalidCharacters(t *testing.T) {
	cases := []struct {
		scenario string
		value    string
	}{
		{"upper", "ABZ"},
		{"upper and lower mixed", "AbZ"},
		{"includes hash", "#faas"},
		{"includes underscore", "test_function"},
		{"ends with hyphen", "testfunction-"},
		{"starts with hyphen", "-testfunction"},
	}

	for _, testCase := range cases {
		request := types.FunctionDeployment{
			Service: testCase.value,
		}

		err := handlers.ValidateDeployRequest(&request)
		if err == nil {
			t.Errorf("Expected error for scenario: %s with value: %s, got: %s", testCase.scenario, testCase.value, err.Error())
		}
	}
}
