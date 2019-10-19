// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package types

import (
	ftypes "github.com/openfaas/faas-provider/types"
)

// ReadConfig constitutes config from env variables
type ReadConfig struct {
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv ftypes.HasEnv) (BootstrapConfig, error) {
	cfg := BootstrapConfig{}

	faasConfig, err := ftypes.ReadConfig{}.Read(hasEnv)
	if err != nil {
		return cfg, err
	}

	cfg.FaaSConfig = *faasConfig

	httpProbe := ftypes.ParseBoolValue(hasEnv.Getenv("http_probe"), false)
	setNonRootUser := ftypes.ParseBoolValue(hasEnv.Getenv("set_nonroot_user"), false)

	readinessProbeInitialDelaySeconds := ftypes.ParseIntValue(hasEnv.Getenv("readiness_probe_initial_delay_seconds"), 3)
	readinessProbeTimeoutSeconds := ftypes.ParseIntValue(hasEnv.Getenv("readiness_probe_timeout_seconds"), 1)
	readinessProbePeriodSeconds := ftypes.ParseIntValue(hasEnv.Getenv("readiness_probe_period_seconds"), 10)

	livenessProbeInitialDelaySeconds := ftypes.ParseIntValue(hasEnv.Getenv("liveness_probe_initial_delay_seconds"), 3)
	livenessProbeTimeoutSeconds := ftypes.ParseIntValue(hasEnv.Getenv("liveness_probe_timeout_seconds"), 1)
	livenessProbePeriodSeconds := ftypes.ParseIntValue(hasEnv.Getenv("liveness_probe_period_seconds"), 10)
	imagePullPolicy := ftypes.ParseString(hasEnv.Getenv("image_pull_policy"), "Always")

	cfg.HTTPProbe = httpProbe
	cfg.SetNonRootUser = setNonRootUser

	cfg.ReadinessProbeInitialDelaySeconds = readinessProbeInitialDelaySeconds
	cfg.ReadinessProbeTimeoutSeconds = readinessProbeTimeoutSeconds
	cfg.ReadinessProbePeriodSeconds = readinessProbePeriodSeconds

	cfg.LivenessProbeInitialDelaySeconds = livenessProbeInitialDelaySeconds
	cfg.LivenessProbeTimeoutSeconds = livenessProbeTimeoutSeconds
	cfg.LivenessProbePeriodSeconds = livenessProbePeriodSeconds

	cfg.ImagePullPolicy = imagePullPolicy

	return cfg, nil
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
	ImagePullPolicy                   string
	FaaSConfig                        ftypes.FaaSConfig
}
