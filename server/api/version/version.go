package version

import "fmt"

var (
	GitDate string
	Version string
	GitHash string
	GitUrl  string
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

func GetGitURL() string {
	hash := GetGitHash()
	if hash == "n/a" {
		hash = "n_a"
	}
	return fmt.Sprintf("https://github.com/freshteapot/learnalist-api/commit/%s", hash)
}
