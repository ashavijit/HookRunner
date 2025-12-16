package version

var (
	Version   = "1.0.0"
	GitCommit = "dev"
	BuildDate = "unknown"
)

func String() string {
	return Version
}

func Full() string {
	return Version + " (" + GitCommit + ") built " + BuildDate
}
