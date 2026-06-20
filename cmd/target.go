package cmd

import (
	"strings"

	"github.com/Omochice/glab-dep/internal/config"
	"github.com/spf13/cobra"
)

// defaultAuthor is the GitLab username searched for when neither --author nor
// dep.author is set. Renovate is the de facto dependency bot on GitLab.
const defaultAuthor = "renovate-bot"

// resolveScope determines the group and project targets.
// Explicit --repo (or dep.repo) wins; otherwise the group filter is used.
// When both are empty, callers search across all accessible projects.
func resolveScope(cmd *cobra.Command, repoValue, groupValue string, cfg *config.Config) (string, []string) {
	var repos []string

	if cmd.Flags().Changed("repo") {
		repos = cleanRepos(repoValue)
	} else if cfg != nil && len(cfg.GetRepos()) > 0 {
		repos = append(repos, cfg.GetRepos()...)
	}

	return groupValue, repos
}

// resolveAuthors picks the effective author filter.
// --author wins, then dep.author, then the default Renovate bot username.
func resolveAuthors(cmd *cobra.Command, authorValue string, cfg *config.Config) []string {
	if cmd.Flags().Changed("author") {
		return []string{authorValue}
	}
	if cfg != nil && cfg.GetAuthor() != "" {
		return []string{cfg.GetAuthor()}
	}
	return []string{defaultAuthor}
}

// cleanRepos splits comma-separated repos, trimming blanks.
func cleanRepos(repoValue string) []string {
	var repos []string
	for r := range strings.SplitSeq(repoValue, ",") {
		r = strings.TrimSpace(r)
		if r != "" {
			repos = append(repos, r)
		}
	}
	return repos
}
