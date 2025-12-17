package version

var (
	Version   = "0.19.0"
	GitCommit = "dev"
	BuildDate = "unknown"
)

func String() string {
	return Version
}

func Full() string {
	return Version + " (" + GitCommit + ") built " + BuildDate
}
