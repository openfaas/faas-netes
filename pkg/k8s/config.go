// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

// ProbeConfig holds the deployment liveness and readiness options
type ProbeConfig struct {
	InitialDelaySeconds int32
	TimeoutSeconds      int32
	PeriodSeconds       int32
}

// DeploymentConfig holds the global deployment options
type DeploymentConfig struct {
	RuntimeHTTPPort int32
	HTTPProbe       bool
	ReadinessProbe  *ProbeConfig
	LivenessProbe   *ProbeConfig
	// SetNonRootUser will override the function image user to ensure that it is not root. When
	// true, the user will set to 12000 for all functions.
	SetNonRootUser bool
}
