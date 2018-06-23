package handlers

import "testing"

func Test_getMinReplicaCount(t *testing.T) {
	scenarios := []struct {
		name   string
		labels *map[string]string
		output int
	}{
		{
			name:   "nil map returns default",
			labels: nil,
			output: initialReplicasCount,
		},
		{
			name:   "empty map returns default",
			labels: &map[string]string{},
			output: initialReplicasCount,
		},
		{
			name:   "empty map returns default",
			labels: &map[string]string{FunctionMinReplicaCount: "2"},
			output: 2,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			output := getMinReplicaCount(s.labels)
			if output == nil {
				t.Errorf("getMinReplicaCount should not return nil pointer")
			}

			value := int(*output)
			if value != s.output {
				t.Errorf("expected: %d, got %d", s.output, value)
			}
		})
	}
}

func Test_parseLabels(t *testing.T) {
	scenarios := []struct {
		name         string
		labels       *map[string]string
		functionName string
		output       map[string]string
	}{
		{
			name:         "nil map returns just the function name",
			labels:       nil,
			functionName: "testFunc",
			output:       map[string]string{FunctionNameLabel: "testFunc"},
		},
		{
			name:         "empty map returns just the function name",
			labels:       &map[string]string{},
			functionName: "testFunc",
			output:       map[string]string{FunctionNameLabel: "testFunc"},
		},
		{
			name:         "non-empty map does not overwrite the openfaas internal labels",
			labels:       &map[string]string{FunctionNameLabel: "anotherValue", "customLabel": "test"},
			functionName: "testFunc",
			output:       map[string]string{FunctionNameLabel: "testFunc", "customLabel": "test"},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			output := buildLabels(s.functionName, s.labels)
			if output == nil {
				t.Errorf("parseLabels should not return nil map")
			}

			outputFuncName := output[FunctionNameLabel]
			if outputFuncName != s.functionName {
				t.Errorf("parseLabels should always set the function name: expected %s, got %s", s.functionName, outputFuncName)
			}

			_, ok := output[FunctionVersionUID]
			if !ok {
				t.Errorf("parseLabels should set the FunctionVersionUID label")
			}

			for key, expectedValue := range s.output {
				value := output[key]
				if value != expectedValue {
					t.Errorf("Incorrect output label for %s, expected: %s, got %s", key, expectedValue, value)
				}
			}

		})
	}
}
