package types

// Reasons an MR cannot be merged, stored in MR.UnmergeableReason. They are
// defined here so the producer (GitLab client) and consumer (TUI labels) share
// one vocabulary and cannot drift apart.
const (
	ReasonConflict   = "conflict"
	ReasonNeedRebase = "need_rebase"
)

// MR represents a merge request
type MR struct {
	IID       int    `json:"iid"`
	ProjectID int    `json:"project_id"` // numeric GitLab project ID, used for API calls
	Title     string `json:"title"`
	Author    string `json:"author"`
	Project   string `json:"project"` // GROUP/PROJECT full path
	URL       string `json:"url"`
	HeadSHA   string `json:"-"`         // head commit SHA
	CIStatus  string `json:"ci_status"` // Pipeline status: success, pending, failure, or empty
	// UnmergeableReason names why the MR currently cannot be merged (for
	// example, "conflict"), or is empty when the MR is mergeable. An unmergeable
	// MR is gated like a failed pipeline rather than offered as a ready-to-merge
	// target.
	UnmergeableReason string `json:"unmergeable_reason"`
}

// Group represents a collection of MRs for the same package@version
type Group struct {
	Key string // package@version
	MRs []MR
}

// Cache represents the cached groups from list --group
type Cache struct {
	Groups map[string][]MR `json:"groups"` // key: package@version, value: list of MRs
}
