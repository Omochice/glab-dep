package tui

import (
	"testing"

	"github.com/Omochice/glab-dep/internal/gitlab"
	"github.com/Omochice/glab-dep/internal/types"
)

func TestFilterMRsExcludesConflictsWhenChecksRequired(t *testing.T) {
	mrs := []types.MR{
		{IID: 1, Project: "group/clean", CIStatus: "success"},
		{IID: 2, Project: "group/conflict", CIStatus: "success", UnmergeableReason: types.ReasonConflict},
	}

	m := NewModel(mrs, "merge", true, ModeMerge, gitlab.SearchParams{}, nil)

	if got := len(m.filteredMRs); got != 1 {
		t.Fatalf("filteredMRs length = %d, want 1", got)
	}
	if m.filteredMRs[0].IID != 1 {
		t.Fatalf("kept MR IID = %d, want 1 (the conflict-free MR)", m.filteredMRs[0].IID)
	}
}

func TestFilterMRsKeepsConflictsWhenChecksNotRequired(t *testing.T) {
	mrs := []types.MR{
		{IID: 1, Project: "group/clean", CIStatus: "success"},
		{IID: 2, Project: "group/conflict", CIStatus: "success", UnmergeableReason: types.ReasonConflict},
	}

	m := NewModel(mrs, "merge", false, ModeMerge, gitlab.SearchParams{}, nil)

	if got := len(m.filteredMRs); got != 2 {
		t.Fatalf("filteredMRs length = %d, want 2", got)
	}
}

func TestFilterMRsExcludesNeedsRebaseWhenChecksRequired(t *testing.T) {
	mrs := []types.MR{
		{IID: 1, Project: "group/clean", CIStatus: "success"},
		{IID: 2, Project: "group/stale", CIStatus: "success", UnmergeableReason: types.ReasonNeedRebase},
	}

	m := NewModel(mrs, "merge", true, ModeMerge, gitlab.SearchParams{}, nil)

	if got := len(m.filteredMRs); got != 1 {
		t.Fatalf("filteredMRs length = %d, want 1", got)
	}
	if m.filteredMRs[0].IID != 1 {
		t.Fatalf("kept MR IID = %d, want 1 (the mergeable MR)", m.filteredMRs[0].IID)
	}
}

func TestUnmergeableLabel(t *testing.T) {
	tests := []struct {
		reason string
		want   string
	}{
		{types.ReasonConflict, "[conflict]"},
		{types.ReasonNeedRebase, "[needs rebase]"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := unmergeableLabel(tt.reason); got != tt.want {
			t.Fatalf("unmergeableLabel(%q) = %q, want %q", tt.reason, got, tt.want)
		}
	}
}
