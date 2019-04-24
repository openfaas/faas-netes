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

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}

	return duration
}

func parseBoolValue(val string, fallback bool) bool {
	if len(val) > 0 {
		return val == "true"
	}
	return fallback
}

func parseString(val string, fallback string) string {
	if len(val) > 0 {
		return val
	}
	return fallback
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv HasEnv) BootstrapConfig {
	cfg := BootstrapConfig{}

	httpProbe := parseBoolValue(hasEnv.Getenv("http_probe"), false)
	setNonRootUser := parseBoolValue(hasEnv.Getenv("set_nonroot_user"), false)

	readinessProbeInitialDelaySeconds := parseIntValue(hasEnv.Getenv("readiness_probe_initial_delay_seconds"), 3)
	readinessProbeTimeoutSeconds := parseIntValue(hasEnv.Getenv("readiness_probe_timeout_seconds"), 1)
	readinessProbePeriodSeconds := parseIntValue(hasEnv.Getenv("readiness_probe_period_seconds"), 10)

	livenessProbeInitialDelaySeconds := parseIntValue(hasEnv.Getenv("liveness_probe_initial_delay_seconds"), 3)
	livenessProbeTimeoutSeconds := parseIntValue(hasEnv.Getenv("liveness_probe_timeout_seconds"), 1)
	livenessProbePeriodSeconds := parseIntValue(hasEnv.Getenv("liveness_probe_period_seconds"), 10)

	readTimeout := parseIntOrDurationValue(hasEnv.Getenv("read_timeout"), time.Second*10)
	writeTimeout := parseIntOrDurationValue(hasEnv.Getenv("write_timeout"), time.Second*10)

	imagePullPolicy := parseString(hasEnv.Getenv("image_pull_policy"), "Always")

	cfg.ReadTimeout = readTimeout
	cfg.WriteTimeout = writeTimeout

	cfg.HTTPProbe = httpProbe
	cfg.SetNonRootUser = setNonRootUser

	cfg.ReadinessProbeInitialDelaySeconds = readinessProbeInitialDelaySeconds
	cfg.ReadinessProbeTimeoutSeconds = readinessProbeTimeoutSeconds
	cfg.ReadinessProbePeriodSeconds = readinessProbePeriodSeconds

	cfg.LivenessProbeInitialDelaySeconds = livenessProbeInitialDelaySeconds
	cfg.LivenessProbeTimeoutSeconds = livenessProbeTimeoutSeconds
	cfg.LivenessProbePeriodSeconds = livenessProbePeriodSeconds

	cfg.ImagePullPolicy = imagePullPolicy

	defaultTCPPort := 8080
	cfg.Port = parseIntValue(hasEnv.Getenv("port"), defaultTCPPort)

	return cfg
}

// BootstrapConfig for the process.
type BootstrapConfig struct {
	// HTTPProbe when set to true switches readiness and liveness probe to
	// access /_/health over HTTP instead of accessing /tmp/.lock.
	HTTPProbe                         bool
	SetNonRootUser                    bool
	ReadinessProbeInitialDelaySeconds int
	ReadinessProbeTimeoutSeconds      int
	ReadinessProbePeriodSeconds       int
	LivenessProbeInitialDelaySeconds  int
	LivenessProbeTimeoutSeconds       int
	LivenessProbePeriodSeconds        int
	ReadTimeout                       time.Duration
	WriteTimeout                      time.Duration
	ImagePullPolicy                   string
	Port                              int
}
