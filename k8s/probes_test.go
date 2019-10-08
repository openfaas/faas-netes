// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"testing"

	types "github.com/openfaas/faas-provider/types"
)

func Test_makeProbes_useExec(t *testing.T) {
	f := mockFactory()

	request := types.FunctionDeployment{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: false,
	}

	probes, err := f.MakeProbes(request)
	if err != nil {
		t.Fatal(err)
	}

	if probes.Readiness.Exec == nil {
		t.Errorf("Readiness probe should have had exec handler")
		t.Fail()
	}
	if probes.Liveness.Exec == nil {
		t.Errorf("Liveness probe should have had exec handler")
		t.Fail()
	}
}

func Test_makeProbes_useHTTPProbe(t *testing.T) {
	f := mockFactory()
	f.Config.HTTPProbe = true

	request := types.FunctionDeployment{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: false,
	}

	probes, err := f.MakeProbes(request)
	if err != nil {
		t.Fatal(err)
	}

	if probes.Readiness.HTTPGet == nil {
		t.Errorf("Readiness probe should have had HTTPGet handler")
		t.Fail()
	}
	if probes.Liveness.HTTPGet == nil {
		t.Errorf("Liveness probe should have had HTTPGet handler")
		t.Fail()
	}
}

func Test_makeProbes_useCustomHTTPProbe(t *testing.T) {
	f := mockFactory()
	customPath := "/healthz"
	request := types.FunctionDeployment{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: false,
		Annotations: &map[string]string{
			ProbePath: customPath,
		},
	}

	probes, err := f.MakeProbes(request)
	if err != nil {
		t.Fatal(err)
	}

	if probes.Readiness.HTTPGet == nil {
		t.Errorf("Readiness probe should have had HTTPGet handler")
		t.Fail()
	}
	if probes.Readiness.HTTPGet.Path != customPath {
		t.Errorf("Readiness probe should have had HTTPGet handler set to %s", customPath)
		t.Fail()
	}

	if probes.Liveness.HTTPGet == nil {
		t.Errorf("Liveness probe should have had HTTPGet handler")
		t.Fail()
	}
	if probes.Liveness.HTTPGet.Path != customPath {
		t.Errorf("Readiness probe should have had HTTPGet handler set to %s", customPath)
		t.Fail()
	}
}

func Test_makeProbes_useCustomDurationHTTPProbe(t *testing.T) {
	f := mockFactory()
	f.Config.HTTPProbe = true
	customDelay := "2m"

	request := types.FunctionDeployment{
		Service:                "testfunc",
		ReadOnlyRootFilesystem: false,
		Annotations: &map[string]string{
			ProbeInitialDelay: customDelay,
		},
	}

	probes, err := f.MakeProbes(request)
	if err != nil {
		t.Fatal(err)
	}

	if probes.Readiness.HTTPGet == nil {
		t.Errorf("Readiness probe should have had HTTPGet handler")
		t.Fail()
	}
	if probes.Readiness.InitialDelaySeconds != 120 {
		t.Errorf("Readiness probe should have initial delay seconds set to %s", customDelay)
		t.Fail()
	}

	if probes.Liveness.HTTPGet == nil {
		t.Errorf("Liveness probe should have had HTTPGet handler")
		t.Fail()
	}
	if probes.Liveness.InitialDelaySeconds != 120 {
		t.Errorf("Readiness probe should have had HTTPGet handler set to %s", customDelay)
		t.Fail()
	}
}
