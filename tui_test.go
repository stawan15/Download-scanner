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
