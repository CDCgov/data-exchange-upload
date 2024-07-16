package version

var (
	GitRepo              string
	LatestReleaseVersion string
	GitShortSha          string
)

type Response struct {
	Repo                 string `json:"repo"`
	LatestReleaseVersion string `json:"latest_release_version"`
	GitShortSha          string `json:"git_short_sha"`
}
