// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package version

var (
	// Version release version of the provider
	Version string

	// GitCommit SHA of the last git commit
	GitCommit string

	// DevVersion string for the development version
	DevVersion = "dev"
)

// BuildVersion returns current version of the provider
func BuildVersion() string {
	if len(Version) == 0 {
		return DevVersion
	}
	return Version
}

// GetReleaseInfo includes the SHA and the release version
func GetReleaseInfo() (sha, release string) {
	return GitCommit, BuildVersion()
}
