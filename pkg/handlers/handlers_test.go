// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright 2020 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"testing"

	types "github.com/openfaas/faas-provider/types"
)

func Test_validateService_ValidCharacters(t *testing.T) {
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

		err := validateService(request.Service)
		if err != nil {
			t.Errorf("Scenario: %s with value: %s, got: %s", testCase.scenario, testCase.value, err.Error())
		}

	}
}

func Test_validateService_InvalidCharacters(t *testing.T) {
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

		err := validateService(request.Service)
		if err == nil {
			t.Errorf("Expected error for scenario: %s with value: %s, got: %s", testCase.scenario, testCase.value, err.Error())
		}
	}
}

func Test_ValidateDeployRequest_InvalidRequest(t *testing.T) {
	cases := []struct {
		scenario    string
		serviceName string
		imageName   string
		err         error
	}{
		{"empty service with valid image",
			"",
			"test-image",
			fmt.Errorf("service: is required")},
		{"empty service with empty image",
			"",
			"",
			fmt.Errorf("service: is required")},
		{"invalid service with valid image",
			"test_function",
			"test-image",
			fmt.Errorf("service: (test_function) is invalid, must be a valid DNS entry")},
		{"invalid service with empty image",
			"test_function",
			"",
			fmt.Errorf("service: (test_function) is invalid, must be a valid DNS entry")},
		{"valid service with empty image",
			"test-function",
			"",
			fmt.Errorf("image: is required")},
	}

	for _, testCase := range cases {
		request := types.FunctionDeployment{
			Service: testCase.serviceName,
			Image:   testCase.imageName,
		}

		err := ValidateDeployRequest(&request)
		if err == nil {
			t.Fatalf("Expected error for scenario: %s", testCase.scenario)
		}
		if err.Error() != testCase.err.Error() {
			t.Fatalf("Scenario: %s, want: %s, but got: %s", testCase.scenario, testCase.err.Error(), err.Error())
		}
	}
}

func Test_ValidateDeploymentRequest_ValidRequest(t *testing.T) {
	request := types.FunctionDeployment{
		Service: "test-function",
		Image:   "test-image",
	}

	err := ValidateDeployRequest(&request)
	if err != nil {
		t.Errorf("unexpected ValidateDeploymentRequest error: %s", err.Error())
	}
}
