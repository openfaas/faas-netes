// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

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
