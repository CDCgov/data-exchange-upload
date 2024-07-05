package version

var (
	GitRepo          string
	LatestReleaseTag string
	GitShortSha      string
)

type Response struct {
	Repo             string `json:"repo"`
	LatestReleaseTag string `json:"latest_release_tag"`
	GitShortSha      string `json:"git_short_sha"`
}
