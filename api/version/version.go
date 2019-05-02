package version

var (
	GitDate string
	Version string
	GitHash string
)

func GetGitDate() string {
	if GitDate == "" {
		return "n/a"
	}
	return GitDate
}

func GetGitHash() string {
	if GitHash == "" {
		return "n/a"
	}
	return GitHash
}

func GetVersion() string {
	if Version == "" {
		return "n/a"
	}
	return Version
}
