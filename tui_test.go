package main

import (
	"strings"
	"testing"
)

func TestTUIShowsWorkspaceSections(t *testing.T) {
	model := newTUIModel(t.TempDir())
	model.width, model.height = 100, 30
	view := model.View()
	for _, want := range []string{"Download Inbox Cleaner", "Overview", "Duplicates", "Organize"} {
		if !strings.Contains(view, want) {
			t.Fatalf("TUI view does not contain %q", want)
		}
	}
}

func TestHumanSize(t *testing.T) {
	if got := humanSize(1536); got != "1.5 KB" {
		t.Fatalf("humanSize() = %q, want 1.5 KB", got)
	}
}

func TestOperationStatusUsesSummaryLine(t *testing.T) {
	output := "Moved: first.png\nMoved: second.png\n\nMoved 2 files. Run 'cleaner undo --yes' to reverse this organization.\n"
	if got := operationStatus(output); !strings.HasPrefix(got, "Moved 2 files.") {
		t.Fatalf("operationStatus() = %q", got)
	}
}
