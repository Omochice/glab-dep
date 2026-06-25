package tui

import (
	"fmt"
	"strings"

	"github.com/Omochice/glab-dep/internal/gitlab"
	"github.com/Omochice/glab-dep/internal/parser"
	"github.com/Omochice/glab-dep/internal/types"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ExecutionMode int

const (
	ModeApprove ExecutionMode = iota
	ModeMerge
	ModeApproveAndMerge
)

func (m ExecutionMode) String() string {
	switch m {
	case ModeApprove:
		return "Approve Only"
	case ModeMerge:
		return "Merge Only"
	case ModeApproveAndMerge:
		return "Approve & Merge"
	default:
		return "Unknown"
	}
}

type ViewState int

const (
	ViewList ViewState = iota
	ViewExecuting
	ViewComplete
	ViewHelp
)

type ExecutionResult struct {
	MR      types.MR
	Action  string
	Success bool
	Error   error
}

type Model struct {
	mrs             []types.MR
	filteredMRs     []types.MR
	selected        map[int]bool // index in filteredMRs
	cursor          int
	mode            ExecutionMode
	view            ViewState
	searchInput     textinput.Model
	searching       bool
	searchQuery     string
	groupFilter     string   // current group filter key (e.g., "lodash@4.17.21")
	customPatterns  []string // custom parsing patterns from config
	executionResult []ExecutionResult
	executing       bool
	refetching      bool
	mergeMethod     string
	requireChecks   bool
	width           int
	height          int
	searchParams    gitlab.SearchParams // For refetching MRs
}

type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Select        key.Binding
	SelectAll     key.Binding
	DeselectAll   key.Binding
	ToggleMode    key.Binding
	ToggleMethod  key.Binding
	ToggleChecks  key.Binding
	Execute       key.Binding
	Search        key.Binding
	GroupFilter   key.Binding
	OpenBrowser   key.Binding
	Refresh       key.Binding
	Help          key.Binding
	Quit          key.Binding
	CancelSearch  key.Binding
	ConfirmSearch key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Select: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("space/enter", "toggle selection"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	),
	DeselectAll: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deselect all"),
	),
	ToggleMode: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "toggle mode"),
	),
	ToggleMethod: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "toggle merge method"),
	),
	ToggleChecks: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "toggle CI checks"),
	),
	Execute: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "execute"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	GroupFilter: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "filter same package"),
	),
	OpenBrowser: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in browser"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh MR list"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	CancelSearch: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel search"),
	),
	ConfirmSearch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm search"),
	),
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	modeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	ciSuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	ciPendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	ciFailureStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	ciUnknownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	conflictStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

func NewModel(mrs []types.MR, mergeMethod string, requireChecks bool, mode ExecutionMode, searchParams gitlab.SearchParams, customPatterns []string) *Model {
	ti := textinput.New()
	ti.Placeholder = "Search MRs..."
	ti.CharLimit = 100

	m := &Model{
		mrs:            mrs,
		filteredMRs:    mrs,
		selected:       make(map[int]bool),
		cursor:         0,
		mode:           mode,
		view:           ViewList,
		searchInput:    ti,
		searching:      false,
		searchQuery:    "",
		groupFilter:    "",
		customPatterns: customPatterns,
		mergeMethod:    mergeMethod,
		requireChecks:  requireChecks,
		searchParams:   searchParams,
	}

	// Apply initial filtering based on requireChecks
	if requireChecks {
		m.filterMRs()
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.view == ViewHelp {
			if msg.String() == "?" || msg.String() == "q" || msg.String() == "esc" {
				m.view = ViewList
			}
			return m, nil
		}

		if m.view == ViewComplete {
			if msg.String() == "q" {
				return m, tea.Quit
			}
			if msg.String() == "enter" || msg.String() == "esc" {
				// Refetch MRs from GitLab to update the list
				m.clearSearch()
				m.view = ViewList
				m.executionResult = nil
				m.selected = make(map[int]bool)
				m.cursor = 0
				m.refetching = true
				return m, m.refetchMRs()
			}
			return m, nil
		}

		if m.executing {
			return m, nil
		}

		if m.refetching {
			return m, nil
		}

		if m.searching {
			switch {
			case key.Matches(msg, keys.ConfirmSearch):
				m.searching = false
				m.searchQuery = m.searchInput.Value()
				m.filterMRs()
				m.cursor = 0
				return m, nil
			case key.Matches(msg, keys.CancelSearch):
				m.searching = false
				m.searchInput.SetValue("")
				m.searchQuery = ""
				m.filterMRs()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Help):
			m.view = ViewHelp
			return m, nil

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.filteredMRs)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.Select):
			m.selected[m.cursor] = !m.selected[m.cursor]

		case key.Matches(msg, keys.SelectAll):
			for i := range m.filteredMRs {
				m.selected[i] = true
			}

		case key.Matches(msg, keys.DeselectAll):
			m.selected = make(map[int]bool)

		case key.Matches(msg, keys.ToggleMode):
			m.mode = (m.mode + 1) % 3

		case key.Matches(msg, keys.ToggleMethod):
			switch m.mergeMethod {
			case "squash":
				m.mergeMethod = "merge"
			case "merge":
				m.mergeMethod = "rebase"
			case "rebase":
				m.mergeMethod = "squash"
			}

		case key.Matches(msg, keys.ToggleChecks):
			m.requireChecks = !m.requireChecks
			m.filterMRs()
			m.cursor = 0

		case key.Matches(msg, keys.Search):
			m.searching = true
			m.searchInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, keys.GroupFilter):
			if len(m.filteredMRs) > 0 && m.cursor < len(m.filteredMRs) {
				currentMR := m.filteredMRs[m.cursor]
				update := parser.ParseTitle(currentMR.Title, m.customPatterns)
				groupKey := update.GroupKey()

				// Toggle: if already filtering by this group, clear it
				if m.groupFilter == groupKey {
					m.groupFilter = ""
				} else {
					m.groupFilter = groupKey
				}
				m.filterMRs()
				m.cursor = 0
			}

		case key.Matches(msg, keys.OpenBrowser):
			if len(m.filteredMRs) > 0 && m.cursor < len(m.filteredMRs) {
				return m, m.openMRInBrowser(m.filteredMRs[m.cursor])
			}

		case key.Matches(msg, keys.Refresh):
			// Clear selections and refetch the MR list
			m.selected = make(map[int]bool)
			m.cursor = 0
			m.refetching = true
			return m, m.refetchMRs()

		case key.Matches(msg, keys.Execute):
			if m.hasSelection() {
				m.executing = true
				m.view = ViewExecuting
				return m, m.executeSelected()
			}
		}

	case ExecutionResult:
		m.executionResult = append(m.executionResult, msg)
		return m, nil

	case executionCompleteMsg:
		m.executing = false
		m.view = ViewComplete
		return m, nil

	case refetchCompleteMsg:
		// Update the MR list with the refetched data
		m.refetching = false
		m.mrs = msg.mrs
		m.filterMRs()
		return m, nil

	case refetchErrorMsg:
		// For now, just continue - we could show an error message in the future
		m.refetching = false
		return m, nil
	}

	return m, nil
}

func (m *Model) View() string {
	switch m.view {
	case ViewHelp:
		return m.renderHelp()
	case ViewExecuting:
		return m.renderExecuting()
	case ViewComplete:
		return m.renderComplete()
	default:
		return m.renderList()
	}
}

func (m *Model) renderList() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("glab-dep Interactive Mode"))
	s.WriteString("\n\n")

	// Show loading indicator if refetching
	if m.refetching {
		s.WriteString(headerStyle.Render("Refreshing MR list from GitLab..."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Please wait..."))
		return s.String()
	}

	// Mode and settings indicator
	s.WriteString(headerStyle.Render("Action: "))
	s.WriteString(modeStyle.Render(m.mode.String()))
	s.WriteString("  ")

	s.WriteString(headerStyle.Render("Method: "))
	s.WriteString(modeStyle.Render(m.mergeMethod))
	s.WriteString("  ")

	if m.requireChecks {
		s.WriteString("  ")
		s.WriteString(headerStyle.Render("Checks: "))
		s.WriteString(modeStyle.Render("required"))
	}

	s.WriteString("\n\n")

	// Search bar
	if m.searching {
		s.WriteString(headerStyle.Render("Search: "))
		s.WriteString(m.searchInput.View())
		s.WriteString("\n\n")
	} else if m.searchQuery != "" {
		s.WriteString(helpStyle.Render(fmt.Sprintf("Filtered by: %q", m.searchQuery)))
		s.WriteString("\n\n")
	}

	// Group filter indicator
	if m.groupFilter != "" {
		s.WriteString(helpStyle.Render(fmt.Sprintf("Group filter: %s (press g to clear)", m.groupFilter)))
		s.WriteString("\n\n")
	}

	// MR list
	s.WriteString(headerStyle.Render(fmt.Sprintf("MRs (%d selected / %d total):", m.countSelected(), len(m.filteredMRs))))
	s.WriteString("\n\n")

	visibleStart, visibleEnd := m.getVisibleRange()

	for i := visibleStart; i < visibleEnd; i++ {
		if i >= len(m.filteredMRs) {
			break
		}

		mr := m.filteredMRs[i]
		cursor := " "
		if i == m.cursor {
			cursor = cursorStyle.Render("❯")
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = selectedStyle.Render("[✓]")
		}

		// A conflicting MR is unmergeable, so it must not read as a ready,
		// green target: replace the pipeline tick with a warning-colored
		// conflict marker rather than a green check.
		status := formatCIStatus(mr.CIStatus)
		title := mr.Title
		if mr.UnmergeableReason != "" {
			status = conflictStyle.Render("⚠")
			title = mr.Title + conflictStyle.Render(" [conflict]")
		}

		line := fmt.Sprintf("%s %s %s %s !%d - %s",
			cursor,
			checkbox,
			status,
			shortenProjectPath(mr.Project),
			mr.IID,
			title,
		)

		if i == m.cursor {
			line = cursorStyle.Render(line)
		}

		s.WriteString(line)
		s.WriteString("\n")
	}

	// Help
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/↓: navigate • space: select • a: select all • d: deselect all • r: refresh"))
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("m/M/c: toggle settings • /: search • g: group • o: open • x: execute • ?: help • q: quit"))

	return s.String()
}

func (m *Model) renderExecuting() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Executing..."))
	s.WriteString("\n\n")

	for _, result := range m.executionResult {
		status := successStyle.Render("✓")
		if !result.Success {
			status = errorStyle.Render("✗")
		}

		msg := fmt.Sprintf("%s %s %s !%d",
			status,
			result.Action,
			result.MR.Project,
			result.MR.IID,
		)

		if !result.Success {
			msg += errorStyle.Render(fmt.Sprintf(" - %v", result.Error))
		}

		s.WriteString(msg)
		s.WriteString("\n")
	}

	if !m.executing {
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Press enter or q to exit"))
	}

	return s.String()
}

func (m *Model) renderComplete() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Execution Complete"))
	s.WriteString("\n\n")

	successCount := 0
	failCount := 0

	for _, result := range m.executionResult {
		status := successStyle.Render("✓")
		if !result.Success {
			status = errorStyle.Render("✗")
			failCount++
		} else {
			successCount++
		}

		msg := fmt.Sprintf("%s %s %s !%d",
			status,
			result.Action,
			result.MR.Project,
			result.MR.IID,
		)

		if !result.Success {
			msg += errorStyle.Render(fmt.Sprintf(" - %v", result.Error))
		}

		s.WriteString(msg)
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(headerStyle.Render(fmt.Sprintf("Summary: %d succeeded, %d failed", successCount, failCount)))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Press enter to return to list • q to quit"))

	return s.String()
}

func (m *Model) renderHelp() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Help - Keyboard Shortcuts"))
	s.WriteString("\n\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Move cursor up"},
		{"↓/j", "Move cursor down"},
		{"space/enter", "Toggle selection of current item"},
		{"a", "Select all visible MRs"},
		{"d", "Deselect all MRs"},
		{"m", "Toggle action mode (Approve → Merge → Approve & Merge)"},
		{"M", "Toggle merge method (squash → merge → rebase)"},
		{"c", "Toggle CI checks requirement"},
		{"/", "Enter search mode"},
		{"g", "Filter by same package@version (toggle)"},
		{"esc", "Cancel search / clear filters"},
		{"o", "Open current MR in browser"},
		{"r", "Refresh MR list from GitLab"},
		{"x", "Execute selected actions"},
		{"?", "Show/hide this help screen"},
		{"q", "Quit the application"},
	}

	for _, sh := range shortcuts {
		fmt.Fprintf(&s, "  %s - %s\n",
			selectedStyle.Render(fmt.Sprintf("%-15s", sh.key)),
			sh.desc,
		)
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Press ? or q to return"))

	return s.String()
}

func (m *Model) hasSelection() bool {
	for _, selected := range m.selected {
		if selected {
			return true
		}
	}
	return false
}

func formatCIStatus(status string) string {
	switch status {
	case "success":
		return ciSuccessStyle.Render("✓")
	case "pending":
		return ciPendingStyle.Render("●")
	case "failure", "error":
		return ciFailureStyle.Render("✗")
	default:
		return ciUnknownStyle.Render("-")
	}
}

func (m *Model) countSelected() int {
	count := 0
	for _, selected := range m.selected {
		if selected {
			count++
		}
	}
	return count
}

func (m *Model) filterMRs() {
	query := strings.ToLower(m.searchQuery)
	var filtered []types.MR

	for _, mr := range m.mrs {
		// Filter by group (package@version) if set
		if m.groupFilter != "" {
			update := parser.ParseTitle(mr.Title, m.customPatterns)
			if update.GroupKey() != m.groupFilter {
				continue
			}
		}

		// Filter by search query if present
		if m.searchQuery != "" {
			matchesSearch := strings.Contains(strings.ToLower(mr.Title), query) ||
				strings.Contains(strings.ToLower(mr.Project), query) ||
				strings.Contains(fmt.Sprintf("%d", mr.IID), query)
			if !matchesSearch {
				continue
			}
		}

		// When checks are required, only keep MRs that are actually mergeable:
		// the pipeline must have succeeded and there must be no conflicts.
		if m.requireChecks && (mr.CIStatus != "success" || mr.UnmergeableReason != "") {
			continue
		}

		filtered = append(filtered, mr)
	}

	m.filteredMRs = filtered

	// Clear selections that are no longer visible
	newSelected := make(map[int]bool)
	m.selected = newSelected
}

func (m *Model) clearSearch() {
	m.searching = false
	m.searchQuery = ""
	m.groupFilter = ""
	m.searchInput.SetValue("")
	m.searchInput.Blur()
}

func (m *Model) getVisibleRange() (int, int) {
	if m.height == 0 {
		return 0, len(m.filteredMRs)
	}

	// Reserve space for header, footer, etc
	maxVisible := max(m.height-15, 5)

	start := max(m.cursor-maxVisible/2, 0)

	end := start + maxVisible
	if end > len(m.filteredMRs) {
		end = len(m.filteredMRs)
		start = max(end-maxVisible, 0)
	}

	return start, end
}

// refetchMRs creates a command to refetch the MR list from GitLab
func (m *Model) refetchMRs() tea.Cmd {
	return func() tea.Msg {
		mrs, err := gitlab.SearchMRs(m.searchParams)
		if err != nil {
			return refetchErrorMsg{err: err}
		}
		return refetchCompleteMsg{mrs: mrs}
	}
}

type executionCompleteMsg struct{}

type refetchCompleteMsg struct {
	mrs []types.MR
}

type refetchErrorMsg struct {
	err error
}
