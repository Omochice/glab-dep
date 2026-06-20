// Package glab invokes the glab CLI. The glab CLI owns the GitLab credentials;
// this tool never stores a token itself.
package glab

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes the glab CLI with the given arguments and returns its stdout.
// On failure the error includes glab's stderr so the underlying cause (for
// example a missing login) is surfaced verbatim.
func Run(args ...string) (string, error) {
	cmd := exec.Command("glab", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("glab %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
