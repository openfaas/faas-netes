// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package test

import (
	"fmt"
	"testing"
	"time"

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

func TestRead_EmptyTimeoutConfig(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := types.ReadConfig{}

	config := readConfig.Read(defaults)
	defaultVal := time.Duration(10) * time.Second
	if (config.ReadTimeout) != defaultVal {
		t.Logf("ReadTimeout want: %s, got %s", defaultVal.String(), config.ReadTimeout.String())
		t.Fail()
	}
	if (config.WriteTimeout) != defaultVal {
		t.Logf("WriteTimeout want: %s, got %s", defaultVal.String(), config.ReadTimeout.String())
		t.Fail()
	}

	wantPort := 8080

	if config.Port != wantPort {
		t.Logf("Port want: %d, got %d", wantPort, config.Port)
		t.Fail()
	}
}

func TestRead_ReadPortConfig(t *testing.T) {
	defaults := NewEnvBucket()
	wantPort := 8082

	defaults.Setenv("port", fmt.Sprintf("%d", wantPort))

	readConfig := types.ReadConfig{}
	config := readConfig.Read(defaults)

	if config.Port != wantPort {
		t.Logf("Port incorrect, want: %d, got: %d\n", wantPort, config.Port)
		t.Fail()
	}
}

func TestRead_ReadAndWriteIntegerTimeoutConfig(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("read_timeout", "10")
	defaults.Setenv("write_timeout", "60")

	readConfig := types.ReadConfig{}
	config := readConfig.Read(defaults)

	if (config.ReadTimeout) != time.Duration(10)*time.Second {
		t.Logf("ReadTimeout incorrect, got: %d\n", config.ReadTimeout)
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(60)*time.Second {
		t.Logf("WriteTimeout incorrect, got: %d\n", config.WriteTimeout)
		t.Fail()
	}
}

func TestRead_ReadAndWriteDurationTimeoutConfig(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("read_timeout", "10s")
	defaults.Setenv("write_timeout", "60s")

	readConfig := types.ReadConfig{}
	config := readConfig.Read(defaults)

	if (config.ReadTimeout) != time.Duration(10)*time.Second {
		t.Logf("ReadTimeout incorrect, got: %d\n", config.ReadTimeout)
		t.Fail()
	}
	if (config.WriteTimeout) != time.Duration(60)*time.Second {
		t.Logf("WriteTimeout incorrect, got: %d\n", config.WriteTimeout)
		t.Fail()
	}
}

func TestRead_EmptyProbeConfig(t *testing.T) {
	defaults := NewEnvBucket()
	readConfig := types.ReadConfig{}
	config := readConfig.Read(defaults)
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
	config := readConfig.Read(defaults)

	if config.HTTPProbe {
		t.Logf("HTTPProbe incorrect, got: %v\n", config.HTTPProbe)
		t.Fail()
	}
}

func TestRead_HTTPProbeConfig_true(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("http_probe", "true")

	readConfig := types.ReadConfig{}
	config := readConfig.Read(defaults)

	if !config.HTTPProbe {
		t.Logf("HTTPProbe incorrect, got: %v\n", config.HTTPProbe)
		t.Fail()
	}
}

func TestRead_ImagePullPolicy_set(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("image_pull_policy", "IfNotPresent")

	readConfig := types.ReadConfig{}
	config := readConfig.Read(defaults)

	if (config.ImagePullPolicy) != "IfNotPresent" {
		t.Logf("ImagePullPolicy incorrect, got: %v\n", config.ImagePullPolicy)
		t.Fail()
	}
}

func TestRead_ImagePullPolicy_empty(t *testing.T) {
	defaults := NewEnvBucket()
	defaults.Setenv("image_pull_policy", "")

	readConfig := types.ReadConfig{}
	config := readConfig.Read(defaults)

	if (config.ImagePullPolicy) != "Always" {
		t.Logf("ImagePullPolicy incorrect, got: %v\n", config.ImagePullPolicy)
		t.Fail()
	}
}
