package version

var (
	Version   = "0.1.1 -beta"
	GitCommit = "dev"
	BuildDate = "unknown"
)

func String() string {
	return Version
}

func Full() string {
	return Version + " (" + GitCommit + ") built " + BuildDate
}
