package tui

import (
	"testing"
)

func TestShortenProjectPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "single segment is kept as-is",
			path: "short-repo",
			want: "short-repo",
		},
		{
			name: "namespace segments are shortened to one character",
			path: "group/subgroup/long-repo-name",
			want: "g/s/long-repo-name",
		},
		{
			name: "two segments shorten only the namespace",
			path: "group/repo",
			want: "g/repo",
		},
		{
			name: "deep path keeps only the tail in full",
			path: "alpha/beta/gamma/delta",
			want: "a/b/g/delta",
		},
		{
			name: "leading dot is preserved with the next character",
			path: ".config/repo",
			want: ".c/repo",
		},
		{
			name: "multibyte segments are shortened by rune, not byte",
			path: "日本語/テスト/リポジトリ",
			want: "日/テ/リポジトリ",
		},
		{
			name: "empty path stays empty",
			path: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shortenProjectPath(tt.path); got != tt.want {
				t.Errorf("shortenProjectPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
