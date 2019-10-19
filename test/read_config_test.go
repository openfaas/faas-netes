// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package test

import (
	"testing"

	"github.com/openfaas/faas-netes/types"
)

type EnvBucket struct {
	Items map[string]string
}

func NewEnvBucket() EnvBucket {
	return EnvBucket{
		Items: make(map[string]string),
	}
}

func (e EnvBucket) Getenv(key string) string {
	return e.Items[key]
}

func (e EnvBucket) Setenv(key string, value string) {
	e.Items[key] = value
}
func TestRead_EmptyProbeConfig(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := types.ReadConfig{}
	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("Unexpected error while reading env %s", err.Error())
	}
	want := false
	if config.HTTPProbe != want {
		t.Logf("EnableFunctionReadinressProbe incorrect, want: %t, got: %t", want, config.HTTPProbe)
		t.Fail()
	}
}

func TestRead_HTTPProbeConfig(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("http_probe", "false")

	readConfig := types.ReadConfig{}
	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("Unexpected error while reading env %s", err.Error())
	}

	if config.HTTPProbe {
		t.Logf("HTTPProbe incorrect, got: %v\n", config.HTTPProbe)
		t.Fail()
	}
}

func TestRead_HTTPProbeConfig_true(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("http_probe", "true")

	readConfig := types.ReadConfig{}
	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("Unexpected error while reading env %s", err.Error())
	}

	if !config.HTTPProbe {
		t.Logf("HTTPProbe incorrect, got: %v\n", config.HTTPProbe)
		t.Fail()
	}
}

func TestRead_ImagePullPolicy_set(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("image_pull_policy", "IfNotPresent")

	readConfig := types.ReadConfig{}
	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("Unexpected error while reading env %s", err.Error())
	}

	if (config.ImagePullPolicy) != "IfNotPresent" {
		t.Logf("ImagePullPolicy incorrect, got: %v\n", config.ImagePullPolicy)
		t.Fail()
	}
}

func TestRead_ImagePullPolicy_empty(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("image_pull_policy", "")

	readConfig := types.ReadConfig{}
	config, err := readConfig.Read(defaults)
	if err != nil {
		t.Fatalf("Unexpected error while reading env %s", err.Error())
	}

	if (config.ImagePullPolicy) != "Always" {
		t.Logf("ImagePullPolicy incorrect, got: %v\n", config.ImagePullPolicy)
		t.Fail()
	}
}
