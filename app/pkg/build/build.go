package build

var (
	version = "0.0.0"
	commit  = "main"
)

// GetVersion returns the semantic release of this build of a service.
func GetVersion() string {
	return version
}

// GetCommit returns the commit reference of this build of a service up to
// n characters. If the commit is shorter than n, the whole commit is returned.
func GetCommit(n int) string {
	if len(commit) > n {
		return commit[:n]
	}

	return commit
}
