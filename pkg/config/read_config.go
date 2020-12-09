// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Author(s) 2020. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package config

import (
	"fmt"
	"log"

	ftypes "github.com/openfaas/faas-provider/types"
)

var validPullPolicyOptions = map[string]bool{
	"Always":       true,
	"IfNotPresent": true,
	"Never":        true,
}

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

	if !validPullPolicyOptions[imagePullPolicy] {
		return cfg, fmt.Errorf("invalid image_pull_policy configured: %s", imagePullPolicy)
	}

	cfg.DefaultFunctionNamespace = ftypes.ParseString(hasEnv.Getenv("function_namespace"), "default")
	cfg.ProfilesNamespace = ftypes.ParseString(hasEnv.Getenv("profiles_namespace"), cfg.DefaultFunctionNamespace)
	cfg.ClusterRole = ftypes.ParseBoolValue(hasEnv.Getenv("cluster_role"), false)

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

// BootstrapConfig contains the server configuration values as well as default
// Function configuration parameters that are passed to the function factory.
type BootstrapConfig struct {
	// HTTPProbe when set to true switches readiness and liveness probe to
	// access /_/health over HTTP instead of accessing /tmp/.lock.
	HTTPProbe bool

	// SetNonRootUser determines if the Function is deployed with a overridden
	// non-root user id.  Currently this is preconfigured to the uid 12000.
	SetNonRootUser bool

	// ReadinessProbeInitialDelaySeconds controls the value of
	// ReadinessProbeInitialDelaySeconds in the Function  ReadinessProbe
	ReadinessProbeInitialDelaySeconds int

	// ReadinessProbeTimeoutSeconds controls the value of
	// ReadinessProbeTimeoutSeconds in the Function  ReadinessProbe
	ReadinessProbeTimeoutSeconds int

	// ReadinessProbePeriodSeconds controls the value of
	// ReadinessProbePeriodSeconds in the Function  ReadinessProbe
	ReadinessProbePeriodSeconds int

	// LivenessProbeInitialDelaySeconds controls the value of
	// LivenessProbeInitialDelaySeconds in the Function  LivenessProbe
	LivenessProbeInitialDelaySeconds int

	// LivenessProbeTimeoutSeconds controls the value of
	// LivenessProbeTimeoutSeconds in the Function  LivenessProbe
	LivenessProbeTimeoutSeconds int

	// LivenessProbePeriodSeconds controls the value of
	// LivenessProbePeriodSeconds in the Function  LivenessProbe
	LivenessProbePeriodSeconds int

	// ImagePullPolicy controls the ImagePullPolicy set on the Function Deployment.
	ImagePullPolicy string

	// DefaultFunctionNamespace defines which namespace in which Functions are deployed.
	// Value is set via the function_namespace environment variable. If the
	// variable is not set, it is set to "default".
	DefaultFunctionNamespace string

	// ProfilesNamespace defines which namespace is used to look up available Profiles.
	// Value is set via the profiles_namespace environment variable. If the
	// variable is not set, then it falls back to DefaultFunctionNamespace.
	ProfilesNamespace string

	// FaaSConfig contains the configuration for the FaaSProvider
	FaaSConfig ftypes.FaaSConfig

	// ClusterRole determines whether the operator should have cluster wide access
	ClusterRole bool
}

// Fprint pretty-prints the config with the stdlib logger. One line per config value.
// When the verbose flag is set to false, it prints the same output as prior to
// the 0.12.0 release.
func (c BootstrapConfig) Fprint(verbose bool) {
	log.Printf("HTTP Read Timeout: %s\n", c.FaaSConfig.GetReadTimeout())
	log.Printf("HTTP Write Timeout: %s\n", c.FaaSConfig.WriteTimeout)
	log.Printf("ImagePullPolicy: %s\n", c.ImagePullPolicy)
	log.Printf("DefaultFunctionNamespace: %s\n", c.DefaultFunctionNamespace)

	if verbose {
		log.Printf("MaxIdleConns: %d\n", c.FaaSConfig.MaxIdleConns)
		log.Printf("MaxIdleConnsPerHost: %d\n", c.FaaSConfig.MaxIdleConnsPerHost)
		log.Printf("HTTPProbe: %v\n", c.HTTPProbe)
		log.Printf("ProfilesNamespace: %s\n", c.ProfilesNamespace)
		log.Printf("SetNonRootUser: %v\n", c.SetNonRootUser)
		log.Printf("ReadinessProbeInitialDelaySeconds: %d\n", c.ReadinessProbeInitialDelaySeconds)
		log.Printf("ReadinessProbeTimeoutSeconds: %d\n", c.ReadinessProbeTimeoutSeconds)
		log.Printf("ReadinessProbePeriodSeconds: %d\n", c.ReadinessProbePeriodSeconds)
		log.Printf("LivenessProbeInitialDelaySeconds: %d\n", c.LivenessProbeInitialDelaySeconds)
		log.Printf("LivenessProbeTimeoutSeconds: %d\n", c.LivenessProbeTimeoutSeconds)
		log.Printf("LivenessProbePeriodSeconds: %d\n", c.LivenessProbePeriodSeconds)
		log.Printf("ClusterRole: %v\n", c.ClusterRole)
	}
}
