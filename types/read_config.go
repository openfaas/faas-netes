// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package types

import (
	"os"
	"strconv"
	"time"
)

// OsEnv implements interface to wrap os.Getenv
type OsEnv struct {
}

// Getenv wraps os.Getenv
func (OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

// HasEnv provides interface for os.Getenv
type HasEnv interface {
	Getenv(key string) string
}

// ReadConfig constitutes config from env variables
type ReadConfig struct {
}

func parseIntValue(val string, fallback int) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return parsedVal
		}
	}
	return fallback
}

func parseBoolValue(val string, fallback bool) bool {
	if len(val) > 0 {
		return val == "true"
	}
	return fallback
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv HasEnv) BootstrapConfig {
	cfg := BootstrapConfig{}

	readTimeout := parseIntValue(hasEnv.Getenv("read_timeout"), 8)
	writeTimeout := parseIntValue(hasEnv.Getenv("write_timeout"), 8)
	enableProbe := parseBoolValue(hasEnv.Getenv("enable_function_readiness_probe"), true)

	cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second
	cfg.EnableFunctionReadinessProbe = enableProbe

	return cfg
}

// BootstrapConfig for the process.
type BootstrapConfig struct {
	EnableFunctionReadinessProbe bool
	ReadTimeout                  time.Duration
	WriteTimeout                 time.Duration
}
