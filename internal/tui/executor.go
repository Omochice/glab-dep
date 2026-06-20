package tui

import (
	"fmt"

	"github.com/Omochice/glab-dep/internal/gitlab"
	"github.com/Omochice/glab-dep/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

// executeSelected returns a command that executes actions on selected MRs
func (m *Model) executeSelected() tea.Cmd {
	// Get list of selected MRs
	var selectedMRs []types.MR
	for i, mr := range m.filteredMRs {
		if m.selected[i] {
			selectedMRs = append(selectedMRs, mr)
		}
	}

	// Return a batch of commands - one for each MR
	var cmds []tea.Cmd
	for _, mr := range selectedMRs {
		cmds = append(cmds, m.executeMRCmd(mr))
	}

	// Run all MR commands concurrently, but only mark completion after they all finish.
	return tea.Sequence(
		tea.Batch(cmds...),
		func() tea.Msg { return executionCompleteMsg{} },
	)
}

// executeMRCmd creates a command to execute action on a single MR
func (m *Model) executeMRCmd(mr types.MR) tea.Cmd {
	return func() tea.Msg {
		switch m.mode {
		case ModeApprove:
			return m.approveMR(mr)
		case ModeMerge:
			return m.mergeMR(mr)
		case ModeApproveAndMerge:
			// First approve
			approveResult := m.approveMR(mr)
			if !approveResult.Success {
				return approveResult
			}
			// Then merge
			return m.mergeMR(mr)
		}
		return ExecutionResult{
			MR:      mr,
			Action:  "unknown",
			Success: false,
			Error:   fmt.Errorf("unknown execution mode"),
		}
	}
}

func (m *Model) approveMR(mr types.MR) ExecutionResult {
	err := gitlab.ApproveMR(mr.Project, mr.IID)
	return ExecutionResult{
		MR:      mr,
		Action:  "approve",
		Success: err == nil,
		Error:   err,
	}
}

func (m *Model) mergeMR(mr types.MR) ExecutionResult {
	// Check CI status if required
	if m.requireChecks {
		headSHA := mr.HeadSHA
		if headSHA == "" {
			sha, err := gitlab.GetMRHead(mr.Project, mr.IID)
			if err != nil {
				return ExecutionResult{
					MR:      mr,
					Action:  "merge (skipped)",
					Success: false,
					Error:   fmt.Errorf("failed to fetch MR head: %w", err),
				}
			}
			headSHA = sha
		}

		status, err := gitlab.GetPipelineStatus(mr.Project, headSHA)
		if err != nil {
			return ExecutionResult{
				MR:      mr,
				Action:  "merge (skipped)",
				Success: false,
				Error:   fmt.Errorf("failed to check CI status: %w", err),
			}
		}

		if !status.AllPassed {
			return ExecutionResult{
				MR:      mr,
				Action:  "merge (skipped)",
				Success: false,
				Error:   fmt.Errorf("CI checks not passing (state: %s)", status.State),
			}
		}
	}

	err := gitlab.MergeMR(mr.Project, mr.IID, m.mergeMethod)
	action := "merge (api)"

	return ExecutionResult{
		MR:      mr,
		Action:  action,
		Success: err == nil,
		Error:   err,
	}
}
