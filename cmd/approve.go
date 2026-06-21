package cmd

import (
	"fmt"

	"github.com/Omochice/glab-dep/internal/cache"
	"github.com/Omochice/glab-dep/internal/gitlab"
	"github.com/Omochice/glab-dep/internal/ui"
	"github.com/spf13/cobra"
)

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Bulk approve all MRs in a group",
	RunE:  runApprove,
}

var (
	approveGroup  string
	approveDryRun bool
)

func init() {
	approveCmd.Flags().StringVar(&approveGroup, "group", "", "Group key (package@version)")
	_ = approveCmd.MarkFlagRequired("group")

	approveCmd.Flags().BoolVar(&approveDryRun, "dry-run", false, "Print actions without executing")
}

func runApprove(cmd *cobra.Command, args []string) error {
	c, err := cache.Load()
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}

	if c == nil || len(c.Groups) == 0 {
		return fmt.Errorf("no cached groups found. Run 'glab dep list --group' first")
	}

	mrs, ok := c.Groups[approveGroup]
	if !ok {
		return fmt.Errorf("group '%s' not found in cache", approveGroup)
	}

	display := ui.New(mrs, false)

	for _, mr := range mrs {
		if approveDryRun {
			display.PrintAction("approve", mr)
			continue
		}

		if err := gitlab.ApproveMR(mr.Project, mr.IID); err != nil {
			display.PrintError("approve", mr, err)
			continue
		}

		display.PrintAction("approve", mr)
	}

	return nil
}
