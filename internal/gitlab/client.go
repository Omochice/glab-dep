package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/Omochice/glab-dep/internal/parser"
	"github.com/Omochice/glab-dep/internal/types"
	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
)

// GetClient returns a GitHub REST API client
func GetClient() (*api.RESTClient, error) {
	return api.DefaultRESTClient()
}

type SearchParams struct {
	Owner           string
	Repos           []string
	Label           string
	Authors         []string
	Limit           int
	ReviewRequested string
	Archived        bool
}

// SearchMRs searches for MRs based on the given parameters.
// When multiple authors are specified, runs one search per author and merges results.
func SearchMRs(params SearchParams) ([]types.MR, error) {
	authors := params.Authors
	if len(authors) == 0 {
		authors = []string{""}
	}

	var allMRs []types.MR
	seen := make(map[string]bool)

	for _, author := range authors {
		mrs, err := searchMRsForAuthor(params, author)
		if err != nil {
			return nil, err
		}
		for _, mr := range mrs {
			key := fmt.Sprintf("%s#%d", mr.Project, mr.IID)
			if !seen[key] {
				seen[key] = true
				allMRs = append(allMRs, mr)
			}
		}
	}

	// Fetch CI status concurrently with worker pool
	const maxWorkers = 10
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)

	for i := range allMRs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			headSHA, err := GetMRHead(allMRs[idx].Project, allMRs[idx].IID)
			if err == nil {
				allMRs[idx].HeadSHA = headSHA
				ciStatus, err := GetPipelineStatus(allMRs[idx].Project, headSHA)
				if err == nil && ciStatus != nil {
					allMRs[idx].CIStatus = ciStatus.State
				}
			}
		}(i)
	}

	wg.Wait()
	return allMRs, nil
}

func searchMRsForAuthor(params SearchParams, author string) ([]types.MR, error) {
	args := []string{"search", "prs", "is:open"}

	if params.Owner != "" {
		args = append(args, "--owner", params.Owner)
	}
	for _, repo := range params.Repos {
		args = append(args, "--repo", repo)
	}

	if params.Label != "" {
		args = append(args, "--label", params.Label)
	}
	if author != "" {
		args = append(args, "--author", author)
	}
	if params.ReviewRequested != "" {
		args = append(args, "--review-requested", params.ReviewRequested)
	}
	if !params.Archived {
		args = append(args, fmt.Sprintf("--archived=%t", params.Archived))
	}
	args = append(args, "--json", "number,title,author,url,repository")
	if params.Limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", params.Limit))
	}

	var rawMRs []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		URL        string `json:"url"`
		Repository struct {
			NameWithOwner string `json:"nameWithOwner"`
		} `json:"repository"`
	}

	stdOut, stdErr, err := gh.Exec(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search MRs: %w\n%s", err, stdErr.String())
	}

	if err := parseJSON(stdOut.String(), &rawMRs); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	mrs := make([]types.MR, len(rawMRs))
	for i, raw := range rawMRs {
		mrs[i] = types.MR{
			IID:     raw.Number,
			Title:   raw.Title,
			Author:  raw.Author.Login,
			Project: raw.Repository.NameWithOwner,
			URL:     raw.URL,
		}
	}

	return mrs, nil
}

// GroupMRs groups MRs by package@version
func GroupMRs(mrs []types.MR, customPatterns []string) map[string][]types.MR {
	groups := make(map[string][]types.MR)

	for _, mr := range mrs {
		update := parser.ParseTitle(mr.Title, customPatterns)
		key := update.GroupKey()
		groups[key] = append(groups[key], mr)
	}

	return groups
}

// ApproveMR approves a merge request
func ApproveMR(repo string, number int) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	body := map[string]string{
		"event": "APPROVE",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("repos/%s/pulls/%d/reviews", repo, number)
	if err := client.Post(path, bytes.NewReader(bodyBytes), nil); err != nil {
		return fmt.Errorf("failed to approve MR #%d: %w", number, err)
	}

	return nil
}

// MergeMR merges an MR via GitHub API
func MergeMR(repo string, number int, method string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	body := map[string]string{
		"merge_method": method,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("repos/%s/pulls/%d/merge", repo, number)
	if err := client.Put(path, bytes.NewReader(bodyBytes), nil); err != nil {
		return fmt.Errorf("failed to merge MR #%d: %w", number, err)
	}

	return nil
}

// CheckStatus represents CI status
type CheckStatus struct {
	State     string // success, pending, failure, error
	AllPassed bool
}

type statusResponse struct {
	State    string          `json:"state"`
	Statuses []statusContext `json:"statuses"`
}

type statusContext struct {
	State string `json:"state"`
}

type checkSuiteResponse struct {
	CheckSuites []checkSuite `json:"check_suites"`
}

type checkSuite struct {
	Status     string  `json:"status"`
	Conclusion *string `json:"conclusion"`
}

// GetMRHead fetches the HEAD SHA for an MR (useful when SearchMRs doesn't return it)
func GetMRHead(repo string, number int) (string, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}

	var pr struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}

	path := fmt.Sprintf("repos/%s/pulls/%d", repo, number)
	if err := client.Get(path, &pr); err != nil {
		return "", fmt.Errorf("failed to get MR #%d: %w", number, err)
	}

	return pr.Head.SHA, nil
}

// GetPipelineStatus checks the CI status for an MR
func GetPipelineStatus(repo string, sha string) (*CheckStatus, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	var suites checkSuiteResponse

	suitePath := fmt.Sprintf("repos/%s/commits/%s/check-suites", repo, sha)
	suitesErr := client.Get(suitePath, &suites)

	var status statusResponse

	statusPath := fmt.Sprintf("repos/%s/commits/%s/status", repo, sha)
	statusErr := client.Get(statusPath, &status)

	if statusErr != nil && suitesErr != nil {
		return nil, fmt.Errorf("failed to get status for %s@%s: status error: %v; check suites error: %v",
			repo, sha, statusErr, suitesErr)
	}

	state := deriveCIState(suites, status)

	return &CheckStatus{
		State:     state,
		AllPassed: state == "success",
	}, nil
}

// parseJSON is a helper to parse JSON strings (gh.Exec returns Bytes that have String() method)
func parseJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

func deriveCIState(
	suites checkSuiteResponse,
	status statusResponse,
) string {
	allCompleted := true
	allSuccess := true

	if len(suites.CheckSuites) > 0 {
		for _, suite := range suites.CheckSuites {
			// ignore queued suites
			if suite.Status == "queued" {
				continue
			}
			if suite.Status != "completed" {
				allCompleted = false
				break
			}
			// check conclusion
			if suite.Conclusion != nil &&
				slices.Contains([]string{"neutral", "skipped", "success"}, *suite.Conclusion) {
				continue
			}
			allSuccess = false
		}
	}

	if len(status.Statuses) > 0 {
		for _, s := range status.Statuses {
			if s.State != "success" {
				allSuccess = false
			}
			if s.State == "pending" {
				allCompleted = false
			}
		}
	}

	if allCompleted {
		if allSuccess {
			return "success"
		} else {
			return "failure"
		}
	}
	return "pending"
}
