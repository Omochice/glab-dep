package tui

import (
	"strconv"

	"github.com/Omochice/glab-dep/internal/glab"
	"github.com/Omochice/glab-dep/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) openMRInBrowser(mr types.MR) tea.Cmd {
	return func() tea.Msg {
		_, _ = glab.Run("mr", "view", strconv.Itoa(mr.IID), "-R", mr.Project, "--web")
		return nil
	}
}
