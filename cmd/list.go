package cmd

import (
	"fmt"

	"github.com/Omochice/glab-dep/internal/cache"
	"github.com/Omochice/glab-dep/internal/config"
	"github.com/Omochice/glab-dep/internal/gitlab"
	"github.com/Omochice/glab-dep/internal/types"
	"github.com/Omochice/glab-dep/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List dependency MRs, optionally grouped by package@version",
	RunE:  runList,
}

var (
	listLabel    string
	listAuthor   string
	listGroup    bool
	listLimit    int
	listRepo     string
	listGroupArg string
	listJSON     bool
	listReviewer string
)

func init() {
	listCmd.Flags().BoolVar(&listGroup, "group", false, "Group MRs by package@version")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")

	listCmd.Flags().IntVar(&listLimit, "limit", 200, "Max MRs to fetch per project")

	// additional filters
	listCmd.Flags().StringVarP(&listRepo, "repo", "R", "", "Target project(s) (GROUP/PROJECT), comma-separated")
	listCmd.Flags().StringVar(&listLabel, "label", "", "MR label to filter")
	listCmd.Flags().StringVar(&listAuthor, "author", "", "MR author username (defaults to the Renovate bot)")
	listCmd.Flags().StringVar(&listGroupArg, "group-path", "", "Target GitLab group/subgroup full path")
	listCmd.Flags().StringVar(&listReviewer, "reviewer", "", "Filter MRs by reviewer username")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	label := listLabel
	authors := resolveAuthors(cmd, listAuthor, cfg)

	group, repos := resolveScope(cmd, listRepo, listGroupArg, cfg)

	searchParams := gitlab.SearchParams{
		Group:    group,
		Repos:    repos,
		Label:    label,
		Authors:  authors,
		Limit:    listLimit,
		Reviewer: listReviewer,
	}

	allMRs, err := gitlab.SearchMRs(searchParams)
	if err != nil {
		return fmt.Errorf("failed to search MRs: %w", err)
	}

	if len(allMRs) == 0 {
		fmt.Println("No dependency MRs found")
		return nil
	}

	if listGroup {
		groups := gitlab.GroupMRs(allMRs, cfg.GetPatterns())

		// Cache the groups
		c := &types.Cache{
			Groups: groups,
		}
		if err := cache.Save(c); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}

		// Display groups
		display := ui.NewFromGroups(groups, listJSON)
		return display.DisplayGroups(groups)
	}

	display := ui.New(allMRs, listJSON)
	return display.DisplayList(allMRs)
}
