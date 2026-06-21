package gitlab

import "testing"

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
