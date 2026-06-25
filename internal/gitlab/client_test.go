package gitlab

import (
	"errors"
	"testing"
)

func TestIsAlreadyApproved(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "401 from a repeat approval",
			err:  errors.New("glab mr approve 5 -R group/proj: exit status 1\nPOST .../approve: 401 {message: 401 Unauthorized}"),
			want: true,
		},
		{
			name: "unrelated failure",
			err:  errors.New("glab mr approve 5 -R group/proj: exit status 1\n404 Not Found"),
			want: false,
		},
		{
			name: "no error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlreadyApproved(tt.err); got != tt.want {
				t.Fatalf("isAlreadyApproved(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestNormalizePipelineStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   string
	}{
		{name: "success", status: "success", want: "success"},
		{name: "skipped counts as success", status: "skipped", want: "success"},
		{name: "failed", status: "failed", want: "failure"},
		{name: "canceled", status: "canceled", want: "failure"},
		{name: "running is pending", status: "running", want: "pending"},
		{name: "manual is pending", status: "manual", want: "pending"},
		{name: "no pipeline stays empty", status: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePipelineStatus(tt.status)
			if got != tt.want {
				t.Fatalf("normalizePipelineStatus(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestParseMRStatus(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantPipeline string
		wantReason   string
	}{
		{
			name:         "mergeable with passing pipeline",
			body:         `{"head_pipeline":{"status":"success"},"has_conflicts":false,"detailed_merge_status":"mergeable"}`,
			wantPipeline: "success",
			wantReason:   "",
		},
		{
			name:         "has_conflicts flag set",
			body:         `{"head_pipeline":{"status":"success"},"has_conflicts":true,"detailed_merge_status":"mergeable"}`,
			wantPipeline: "success",
			wantReason:   "conflict",
		},
		{
			name:         "detailed_merge_status reports conflict",
			body:         `{"head_pipeline":{"status":"success"},"has_conflicts":false,"detailed_merge_status":"conflict"}`,
			wantPipeline: "success",
			wantReason:   "conflict",
		},
		{
			name:         "detailed_merge_status needs rebase",
			body:         `{"head_pipeline":{"status":"success"},"has_conflicts":false,"detailed_merge_status":"need_rebase"}`,
			wantPipeline: "success",
			wantReason:   "need_rebase",
		},
		{
			name:         "missing fields default to mergeable",
			body:         `{"head_pipeline":{"status":"running"}}`,
			wantPipeline: "pending",
			wantReason:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMRStatus([]byte(tt.body))
			if err != nil {
				t.Fatalf("parseMRStatus returned error: %v", err)
			}
			if got.Pipeline != tt.wantPipeline {
				t.Fatalf("Pipeline = %q, want %q", got.Pipeline, tt.wantPipeline)
			}
			if got.UnmergeableReason != tt.wantReason {
				t.Fatalf("UnmergeableReason = %q, want %q", got.UnmergeableReason, tt.wantReason)
			}
		})
	}
}

func TestProjectPathFromURL(t *testing.T) {
	tests := []struct {
		name   string
		webURL string
		want   string
	}{
		{
			name:   "simple project",
			webURL: "https://gitlab.com/group/project/-/merge_requests/12",
			want:   "group/project",
		},
		{
			name:   "subgroup project",
			webURL: "https://gitlab.example.com/group/sub/project/-/merge_requests/3",
			want:   "group/sub/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := projectPathFromURL(tt.webURL)
			if got != tt.want {
				t.Fatalf("projectPathFromURL(%q) = %q, want %q", tt.webURL, got, tt.want)
			}
		})
	}
}
