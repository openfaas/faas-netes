package version

var (
	//Version release version of the current provider
	Version string
	//GitCommit SHA of the last git commit
	GitCommit string
	//DevVerison string for the development version
	DevVerison = "dev"
	//LastVersion last version of the provider
	LastVersion string
)

//BuildVersion returns current version of the provider
func BuildVersion() string {
	if len(Version) == 0 {
		return LastVersion + "-" + DevVerison
	}
	return Version
}
