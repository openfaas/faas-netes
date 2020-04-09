package version

var SHA string
var Release string

func GetReleaseInfo() (sha, release string) {
	sha = "local-dev"
	release = "latest-dev"

	if len(SHA) > 0 {
		sha = SHA
	}

	if len(Release) > 0 {
		release = Release
	}

	return sha, release
}
