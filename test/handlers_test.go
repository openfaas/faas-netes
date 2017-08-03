package test

import (
	"testing"

	"github.com/alexellis/faas-netes/handlers"
	"github.com/alexellis/faas/gateway/requests"
)

func Test_ValidateDeployRequest_ValidCharacters(t *testing.T) {
	cases := []struct {
		scenario string
		value    string
	}{
		{"lower", "abz"},
		{"upper", "ABZ"},
		{"upper and lower mixed", "AbZ"},
		{"includes dashes", "test-function"},
	}

	for _, testCase := range cases {
		request := requests.CreateFunctionRequest{
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
		{"includes hash", "#faas"},
		{"includes underscore", "test_function"},
	}

	for _, testCase := range cases {
		request := requests.CreateFunctionRequest{
			Service: testCase.value,
		}

		err := handlers.ValidateDeployRequest(&request)
		if err == nil {
			t.Errorf("Expected error for scenario: %s with value: %s, got: %s", testCase.scenario, testCase.value, err.Error())
		}
	}
}
