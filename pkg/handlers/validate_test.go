package handlers

import (
	"fmt"
	"testing"

	"github.com/openfaas/faas-provider/types"
)

func Test_validateScalingLabels(t *testing.T) {

	testCases := []struct {
		Name   string
		Labels map[string]string
		Err    error
	}{
		{
			Name:   "empty labels",
			Labels: map[string]string{},
			Err:    nil,
		},
		{
			Name: "com.openfaas.scale.zero",
			Labels: map[string]string{
				"com.openfaas.scale.zero": "true",
			},
			Err: fmt.Errorf("com.openfaas.scale.zero not available for Community Edition"),
		},
		{
			Name: "com.openfaas.scale.zero-duration",
			Labels: map[string]string{
				"com.openfaas.scale.zero-duration": "15m",
			},
			Err: fmt.Errorf("com.openfaas.scale.zero-duration not available for Community Edition"),
		},
		{
			Name: "com.openfaas.scale.target",
			Labels: map[string]string{
				"com.openfaas.scale.target": "10",
			},
			Err: fmt.Errorf("com.openfaas.scale.target not available for Community Edition"),
		},
		{
			Name: "com.openfaas.scale.type",
			Labels: map[string]string{
				"com.openfaas.scale.type": "cpu",
			},
			Err: fmt.Errorf("com.openfaas.scale.type not available for Community Edition"),
		},
		{
			Name: "com.openfaas.scale.min allowed",
			Labels: map[string]string{
				"com.openfaas.scale.min": "1",
			},
		},
		{
			Name: "com.openfaas.scale.min allowed within range",
			Labels: map[string]string{
				"com.openfaas.scale.min": "5",
			},
		},
		{
			Name: "com.openfaas.scale.max allowed within range",
			Labels: map[string]string{
				"com.openfaas.scale.max": "5",
			},
		},
		{
			Name: "com.openfaas.scale.min allowed within range",
			Labels: map[string]string{
				"com.openfaas.scale.min": "5",
			},
		},
		{
			Name: "com.openfaas.scale.max fails outside of range",
			Labels: map[string]string{
				"com.openfaas.scale.max": "6",
			},
			Err: fmt.Errorf("com.openfaas.scale.max is set too high for Community Edition"),
		},
		{
			Name: "com.openfaas.scale.min fails outside of range",
			Labels: map[string]string{
				"com.openfaas.scale.min": "6",
			},
			Err: fmt.Errorf("com.openfaas.scale.min is set too high for Community Edition"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			gotErr := validateScalingLabels(&types.FunctionDeployment{Labels: &tc.Labels})
			got := fmt.Errorf("")
			if gotErr != nil {
				got = gotErr
			}
			want := fmt.Errorf("")
			if tc.Err != nil {
				want = tc.Err
			}

			if got.Error() != want.Error() {
				t.Errorf("got: %v, want: %v", got, want)
			}
		})
	}

}
