package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryOptions(t *testing.T) {
	dir := t.TempDir()
	got, format, err := directoryOptions([]string{"--dir", dir, "--format", "json"})
	if err != nil || got != dir || format != jsonFormat {
		t.Fatalf("directoryOptions() = %q, %q, %v", got, format, err)
	}
	if _, _, err := directoryOptions([]string{"--format", "xml"}); err == nil {
		t.Fatal("expected invalid output format to fail")
	}
}

func TestScanAndDuplicateResultsAreDeterministic(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string]string{"b.txt": "same", "a.txt": "same", "large.bin": "123456"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "folder"), 0755); err != nil {
		t.Fatal(err)
	}
	scan, err := scanResult(dir)
	if err != nil {
		t.Fatal(err)
	}
	if scan.Files[0].Name != "large.bin" || scan.Directories[0] != "folder" || len(scan.Extensions) != 2 {
		t.Fatalf("unexpected scan result: %#v", scan)
	}
	duplicates, err := duplicateResult(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(duplicates.Groups) != 1 || duplicates.Groups[0][0].Name != "a.txt" || duplicates.ReclaimableBytes != 4 {
		t.Fatalf("unexpected duplicate result: %#v", duplicates)
	}
}

func TestMutationOptionsRequireExplicitFlags(t *testing.T) {
	dir := t.TempDir()
	gotDir, apply, yes, names, err := mutationOptions([]string{"--dir", dir, "--apply", "--yes", "file.txt"}, true)
	if err != nil || gotDir != dir || !apply || !yes || len(names) != 1 {
		t.Fatalf("unexpected mutation options: %q %t %t %#v %v", gotDir, apply, yes, names, err)
	}
	if _, apply, yes, _, err := mutationOptions([]string{"--yes"}, false); err != nil || apply || !yes {
		t.Fatalf("unexpected undo options: %t %t %v", apply, yes, err)
	}
}

func TestLegacyOrganizeApplyAliasStillRequiresYes(t *testing.T) {
	dir := t.TempDir()
	args := []string{"apply", "--dir", dir}
	if args[0] == "apply" {
		args[0] = "--apply"
	}
	_, apply, yes, names, err := mutationOptions(args, true)
	if err != nil || !apply || yes || len(names) != 0 {
		t.Fatalf("legacy apply alias parsed incorrectly: apply=%t yes=%t names=%v err=%v", apply, yes, names, err)
	}
}
