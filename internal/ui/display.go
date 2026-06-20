package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Omochice/glab-dep/internal/types"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/term"
)

// UI encapsulates display state and configuration
type UI struct {
	multiProject bool
	json         bool
}

// New creates a new UI instance
// Automatically detects if MRs span multiple projects
func New(mrs []types.MR, json bool) *UI {
	return &UI{
		multiProject: isMultiProject(mrs),
		json:         json,
	}
}

// NewFromGroups creates a new UI instance from grouped MRs
func NewFromGroups(groups map[string][]types.MR, json bool) *UI {
	return &UI{
		multiProject: isMultiProjectGroups(groups),
		json:         json,
	}
}

// DisplayList prints MRs in a flat list format (JSON or table-like)
func (u *UI) DisplayList(mrs []types.MR) error {
	if u.json {
		return u.displayListJSON(mrs)
	}

	if len(mrs) == 0 {
		return nil
	}

	isTTY := term.IsTerminal(os.Stdout)
	termWidth, _, _ := term.FromEnv().Size()
	table := tableprinter.New(os.Stdout, isTTY, termWidth)

	table.AddHeader([]string{"PROJECT", "MR", "TITLE"})
	for _, mr := range mrs {
		table.AddField(mr.Project)
		table.AddField("!" + strconv.Itoa(mr.IID))
		table.AddField(mr.Title)
		table.EndRow()
	}

	return table.Render()
}

// DisplayGroups prints MRs grouped by package@version (JSON or hierarchical format)
func (u *UI) DisplayGroups(groups map[string][]types.MR) error {
	if u.json {
		return u.displayGroupsJSON(groups)
	}

	isTTY := term.IsTerminal(os.Stdout)
	termWidth, _, _ := term.FromEnv().Size()

	// Sort groups alphabetically for consistent output
	sortedKeys := make([]string, 0, len(groups))
	for k := range groups {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// Create single table for all groups
	table := tableprinter.New(os.Stdout, isTTY, termWidth)
	table.AddHeader([]string{"GROUP", "PROJECT", "MR", "URL"})

	for _, key := range sortedKeys {
		groupMRs := groups[key]

		// Sort MRs within group by project name, then by MR IID
		sort.Slice(groupMRs, func(i, j int) bool {
			if groupMRs[i].Project != groupMRs[j].Project {
				return groupMRs[i].Project < groupMRs[j].Project
			}
			return groupMRs[i].IID < groupMRs[j].IID
		})

		for i, mr := range groupMRs {
			// Group name only on first row of each group
			if i == 0 {
				table.AddField(key)
			} else {
				table.AddField("")
			}

			projectParts := strings.Split(mr.Project, "/")
			projectShort := projectParts[len(projectParts)-1]
			table.AddField(projectShort)

			table.AddField("!" + strconv.Itoa(mr.IID))
			table.AddField(mr.URL)
			table.EndRow()
		}
	}

	return table.Render()
}

// PrintAction prints a standardized action message for an MR
// Examples:
//   - approve !123
//   - [group/project] approve !123
//   - merge !123 via API
//   - [group/project] skipped !123: CI checks not passing
func (u *UI) PrintAction(action string, mr types.MR, details ...string) {
	prefix := ""
	if u.multiProject {
		prefix = fmt.Sprintf("[%s] ", mr.Project)
	}

	message := fmt.Sprintf("%s !%d", action, mr.IID)
	if len(details) > 0 {
		message += ": " + details[0]
	}

	fmt.Printf("%s%s\n", prefix, message)
}

// PrintError prints a standardized error message for an MR action
func (u *UI) PrintError(action string, mr types.MR, err error) {
	prefix := ""
	if u.multiProject {
		prefix = fmt.Sprintf("[%s] ", mr.Project)
	}

	fmt.Printf("%sfailed to %s !%d: %v\n", prefix, action, mr.IID, err)
}

func isMultiProject(mrs []types.MR) bool {
	if len(mrs) == 0 {
		return false
	}

	firstProject := mrs[0].Project
	for _, mr := range mrs {
		if mr.Project != firstProject {
			return true
		}
	}
	return false
}

func isMultiProjectGroups(groups map[string][]types.MR) bool {
	for _, mrs := range groups {
		if len(mrs) > 0 {
			firstProject := mrs[0].Project
			for _, mr := range mrs {
				if mr.Project != firstProject {
					return true
				}
			}
		}
	}
	return false
}

func (u *UI) displayListJSON(mrs []types.MR) error {
	data, err := json.MarshalIndent(mrs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func (u *UI) displayGroupsJSON(groups map[string][]types.MR) error {
	data, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
