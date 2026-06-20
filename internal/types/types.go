package types

// MR represents a merge request
type MR struct {
	IID      int    `json:"iid"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Project  string `json:"project"` // GROUP/PROJECT full path
	URL      string `json:"url"`
	HeadSHA  string `json:"-"`         // For pipeline status checks
	CIStatus string `json:"ci_status"` // Pipeline status: success, pending, failure, or empty
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
