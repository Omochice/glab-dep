package config

import (
	"strings"

	"github.com/Omochice/glab-dep/internal/glab"
)

// Config holds all configuration values read from glab config.
type Config struct {
	Repos    []string // dep.repo (comma-separated)
	Patterns []string // dep.patterns (comma-separated regex patterns)
	Author   string   // dep.author (default dependency bot username)
}

// Load reads configuration from glab config.
// Returns a Config with zero values when nothing is set.
func Load() (*Config, error) {
	cfg := &Config{
		Repos:    splitList(get("dep.repo")),
		Patterns: splitList(get("dep.patterns")),
		Author:   strings.TrimSpace(get("dep.author")),
	}
	return cfg, nil
}

// get reads a single glab config key, returning an empty string when unset
// or when glab is unavailable (config is optional, so errors are not fatal).
func get(key string) string {
	out, err := glab.Run("config", "get", key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func splitList(value string) []string {
	var items []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

// GetRepos returns the configured repos or nil if not set
func (c *Config) GetRepos() []string {
	return c.Repos
}

// GetPatterns returns the configured patterns or nil if not set
func (c *Config) GetPatterns() []string {
	return c.Patterns
}

// GetAuthor returns the configured default author or an empty string if not set
func (c *Config) GetAuthor() string {
	return c.Author
}
