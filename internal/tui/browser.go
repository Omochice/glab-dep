package tui

import (
	"github.com/Omochice/glab-dep/internal/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2"
)

func (m Model) openMRInBrowser(mr types.MR) tea.Cmd {
	return func() tea.Msg {
		args := []string{"pr", "view", mr.URL, "--web"}
		_, _, _ = gh.Exec(args...)
		return nil
	}
}
