// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Author(s) 2020. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package config

import (
	"log"

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

	cfg.DefaultFunctionNamespace = ftypes.ParseString(hasEnv.Getenv("function_namespace"), "openfaas-fn")
	cfg.ProfilesNamespace = ftypes.ParseString(hasEnv.Getenv("profiles_namespace"), cfg.DefaultFunctionNamespace)

	cfg.HTTPProbe = httpProbe
	cfg.SetNonRootUser = setNonRootUser

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
}

// Fprint pretty-prints the config with the stdlib logger. One line per config value.
// When the verbose flag is set to false, it prints the same output as prior to
// the 0.12.0 release.
func (c BootstrapConfig) Fprint(verbose bool) {
	log.Printf("HTTP Read Timeout: %s\n", c.FaaSConfig.GetReadTimeout())
	log.Printf("HTTP Write Timeout: %s\n", c.FaaSConfig.WriteTimeout)

	log.Printf("ImagePullPolicy: %s\n", "Always")
	log.Printf("DefaultFunctionNamespace: %s\n", c.DefaultFunctionNamespace)

	if verbose {
		log.Printf("MaxIdleConns: %d\n", c.FaaSConfig.MaxIdleConns)
		log.Printf("MaxIdleConnsPerHost: %d\n", c.FaaSConfig.MaxIdleConnsPerHost)
		log.Printf("HTTPProbe: %v\n", c.HTTPProbe)
		log.Printf("ProfilesNamespace: %s\n", c.ProfilesNamespace)
		log.Printf("SetNonRootUser: %v\n", c.SetNonRootUser)
	}
}
