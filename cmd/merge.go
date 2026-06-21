package cmd

import (
	"fmt"

	"github.com/Omochice/glab-dep/internal/cache"
	"github.com/Omochice/glab-dep/internal/gitlab"
	"github.com/Omochice/glab-dep/internal/ui"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Bulk merge all MRs in a group",
	RunE:  runMerge,
}

var (
	mergeGroup         string
	mergeDryRun        bool
	mergeMethod        string
	mergeRequireChecks bool
)

func init() {
	mergeCmd.Flags().StringVar(&mergeGroup, "group", "", "Group key (package@version)")
	_ = mergeCmd.MarkFlagRequired("group")

	mergeCmd.Flags().BoolVar(&mergeDryRun, "dry-run", false, "Print actions without executing")
	mergeCmd.Flags().StringVar(&mergeMethod, "method", "squash", "Merge method: merge, squash, or rebase")
	mergeCmd.Flags().BoolVar(&mergeRequireChecks, "require-checks", true, "Merge only when the pipeline succeeds (GitLab auto-merge)")
}

func runMerge(cmd *cobra.Command, args []string) error {
	if mergeMethod != "merge" && mergeMethod != "squash" && mergeMethod != "rebase" {
		return fmt.Errorf("invalid merge method: %s (must be 'merge', 'squash', or 'rebase')", mergeMethod)
	}

	c, err := cache.Load()
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}

	if c == nil || len(c.Groups) == 0 {
		return fmt.Errorf("no cached groups found. Run 'glab dep list --group' first")
	}

	mrs, ok := c.Groups[mergeGroup]
	if !ok {
		return fmt.Errorf("group '%s' not found in cache", mergeGroup)
	}

	display := ui.New(mrs, false)

	for _, mr := range mrs {
		if mergeDryRun {
			display.PrintAction("[dry-run] merge", mr)
			continue
		}

		// With --require-checks, GitLab merges the MR once its pipeline
		// succeeds (native auto-merge) rather than us gating it here.
		mergeErr := gitlab.MergeMR(mr.Project, mr.IID, mergeMethod, mergeRequireChecks)
		if mergeErr != nil {
			display.PrintError("merge", mr, mergeErr)
			continue
		}

		if mergeRequireChecks {
			display.PrintAction("merge", mr, "auto-merge when pipeline succeeds")
		} else {
			display.PrintAction("merge", mr)
		}
	}

	return nil
}
