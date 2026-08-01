package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestTUIQuitsWithoutChangingFiles(t *testing.T) {
	var output bytes.Buffer
	if err := runTUI(strings.NewReader("q\n"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Download Inbox Cleaner") || !strings.Contains(output.String(), "Goodbye") {
		t.Fatalf("unexpected TUI output: %q", output.String())
	}
}

func TestTUIConfirmationRequiresExactWord(t *testing.T) {
	var output bytes.Buffer
	confirmed, err := confirmTUIAction(bufio.NewReader(strings.NewReader("move\n")), &output, "Confirm", "MOVE")
	if err != nil || confirmed {
		t.Fatalf("confirmation = %t, %v; want false, nil", confirmed, err)
	}
}
