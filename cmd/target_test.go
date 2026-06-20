package cmd

import (
	"slices"
	"testing"

	"github.com/Omochice/glab-dep/internal/config"
	"github.com/spf13/cobra"
)

func TestResolveScopeDefaultsToAllProjects(t *testing.T) {
	c := newTestCommand()
	cfg := &config.Config{}

	group, repos := resolveScope(c, "", "", cfg)

	if group != "" {
		t.Fatalf("expected group to remain empty (search all), got %q", group)
	}
	if len(repos) != 0 {
		t.Fatalf("expected no repos, got %v", repos)
	}
}

func TestResolveScopePrefersFlaggedRepos(t *testing.T) {
	c := newTestCommand()
	if err := c.Flags().Set("repo", "group/a, group/sub/b"); err != nil {
		t.Fatalf("failed to set repo flag: %v", err)
	}

	group, repos := resolveScope(c, "group/a, group/sub/b", "", &config.Config{})

	if group != "" {
		t.Fatalf("expected group to remain empty, got %q", group)
	}

	expected := []string{"group/a", "group/sub/b"}
	if !slices.Equal(repos, expected) {
		t.Fatalf("expected repos %v, got %v", expected, repos)
	}
}

func TestResolveScopeUsesConfigRepos(t *testing.T) {
	c := newTestCommand()
	cfg := &config.Config{
		Repos: []string{"group/project"},
	}

	group, repos := resolveScope(c, "", "", cfg)

	if group != "" {
		t.Fatalf("expected group to remain empty when repos are configured, got %q", group)
	}
	if !slices.Equal(repos, cfg.Repos) {
		t.Fatalf("expected repos %v, got %v", cfg.Repos, repos)
	}
}

func TestResolveScopeKeepsExplicitGroup(t *testing.T) {
	c := newTestCommand()
	group, repos := resolveScope(c, "", "mygroup/sub", &config.Config{})

	if group != "mygroup/sub" {
		t.Fatalf("expected group to be %q, got %q", "mygroup/sub", group)
	}
	if len(repos) != 0 {
		t.Fatalf("expected no repos, got %v", repos)
	}
}

func TestResolveAuthorsDefaultsToRenovateBot(t *testing.T) {
	c := newTestCommand()
	authors := resolveAuthors(c, "", &config.Config{})
	expected := []string{defaultAuthor}
	if !slices.Equal(authors, expected) {
		t.Fatalf("expected %v, got %v", expected, authors)
	}
}

func TestResolveAuthorsUsesConfigAuthor(t *testing.T) {
	c := newTestCommand()
	cfg := &config.Config{Author: "my-renovate-bot"}
	authors := resolveAuthors(c, "", cfg)
	expected := []string{"my-renovate-bot"}
	if !slices.Equal(authors, expected) {
		t.Fatalf("expected %v, got %v", expected, authors)
	}
}

func TestResolveAuthorsAuthorFlagOverrides(t *testing.T) {
	c := newTestCommand()
	if err := c.Flags().Set("author", "someuser"); err != nil {
		t.Fatalf("failed to set author flag: %v", err)
	}
	authors := resolveAuthors(c, "someuser", &config.Config{Author: "ignored"})
	expected := []string{"someuser"}
	if !slices.Equal(authors, expected) {
		t.Fatalf("expected %v, got %v", expected, authors)
	}
}

func newTestCommand() *cobra.Command {
	c := &cobra.Command{}
	c.Flags().String("repo", "", "")
	c.Flags().String("author", "", "")
	return c
}
