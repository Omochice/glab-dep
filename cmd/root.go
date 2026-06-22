package cmd

import (
	"fmt"

	"github.com/Omochice/glab-dep/internal/config"
	"github.com/Omochice/glab-dep/internal/gitlab"
	"github.com/Omochice/glab-dep/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	rootLabel        string
	rootAuthor       string
	rootLimit        int
	rootRepo         string
	rootGroup        string
	rootMergeMethod  string
	rootRequireCheck bool
	rootMode         string
	rootReviewer     string
)

var rootCmd = &cobra.Command{
	Use:   "glab-dep",
	Short: "Streamline dependency MR review and merge workflow",
	Long: `glab-dep helps you manage automated dependency update MRs on GitLab
by grouping, bulk approving, and bulk merging them. It uses the glab CLI for
all GitLab access, so it never stores a token itself.

When run without subcommands, launches interactive TUI mode.`,
	SilenceUsage: true,
	RunE:         runRoot,
}

func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	group, repos := resolveScope(cmd, rootRepo, rootGroup, cfg)
	authors := resolveAuthors(cmd, rootAuthor, cfg)

	searchParams := gitlab.SearchParams{
		Group:    group,
		Repos:    repos,
		Label:    rootLabel,
		Authors:  authors,
		Limit:    rootLimit,
		Reviewer: rootReviewer,
	}

	allMRs, err := gitlab.SearchMRs(searchParams)
	if err != nil {
		return fmt.Errorf("failed to search MRs: %w", err)
	}

	if len(allMRs) == 0 {
		fmt.Println("No dependency MRs found")
		return nil
	}

	// Parse mode
	mode := tui.ModeApprove // default
	switch rootMode {
	case "approve":
		mode = tui.ModeApprove
	case "merge":
		mode = tui.ModeMerge
	case "approve-and-merge", "both":
		mode = tui.ModeApproveAndMerge
	}

	// Launch TUI
	model := tui.NewModel(allMRs, rootMergeMethod, rootRequireCheck, mode, searchParams, cfg.GetPatterns())

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

func init() {
	rootCmd.Flags().IntVar(&rootLimit, "limit", 200, "Max MRs to fetch per project")
	rootCmd.Flags().StringVar(&rootLabel, "label", "", "MR label to filter")
	rootCmd.Flags().StringVar(&rootAuthor, "author", "", "MR author username (defaults to the Renovate bot)")
	rootCmd.Flags().StringVarP(&rootRepo, "repo", "R", "", "Target project(s) (GROUP/PROJECT), comma-separated")
	rootCmd.Flags().StringVar(&rootGroup, "group-path", "", "Target GitLab group/subgroup full path")
	rootCmd.Flags().StringVar(&rootMergeMethod, "merge-method", "squash", "Merge method: merge, squash, or rebase")
	rootCmd.Flags().BoolVar(&rootRequireCheck, "require-checks", false, "Merge only when the pipeline succeeds (GitLab auto-merge)")
	rootCmd.Flags().StringVar(&rootMode, "mode", "approve", "Execution mode: approve, merge, or approve-and-merge (both)")
	rootCmd.Flags().StringVar(&rootReviewer, "reviewer", "", "Filter MRs by reviewer username")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(approveCmd)
	rootCmd.AddCommand(mergeCmd)
}
