package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOrganizationFolder(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"report.pdf", "Documents/PDF"},
		{"budget.xlsx", "Documents/Office"},
		{"photo.JPG", "Images"},
		{"portrait.heic", "Images"},
		{"recording.mov", "Media/Video"},
		{"podcast.mp3", "Media/Audio"},
		{"backup.tar.gz", "Archives"},
		{"bundle.7z", "Archives"},
		{"installer.dmg", "Installers"},
		{"font.woff2", "Fonts"},
		{"diagram.excalidraw", "Design"},
		{"notes.txt", "Documents/Text"},
		{"CalendarInfoUC.cs", "Code"},
		{"brakeman.html", "Code"},
		{"unknown.bin", "Other"},
		{"README", "Other"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := organizationFolder(test.name); got != test.want {
				t.Fatalf("organizationFolder(%q) = %q, want %q", test.name, got, test.want)
			}
		})
	}
}

func TestSecurityFinding(t *testing.T) {
	tests := []struct {
		name      string
		wantFound bool
		wantScore int
	}{
		{"invoice.pdf.exe", true, 80},
		{"install.dmg", true, 10},
		{".hidden.sh", true, 55},
		{"notes.txt", false, 0},
		{".DS_Store", false, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			finding, found := securityFinding(test.name)
			if found != test.wantFound {
				t.Fatalf("securityFinding(%q) found = %t, want %t", test.name, found, test.wantFound)
			}
			if finding.Score != test.wantScore {
				t.Fatalf("securityFinding(%q) score = %d, want %d", test.name, finding.Score, test.wantScore)
			}
		})
	}
}

func TestResolveTargetDirectory(t *testing.T) {
	tempDir := t.TempDir()

	got, err := resolveTargetDirectory([]string{tempDir})
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("resolveTargetDirectory() = %q, want %q", got, want)
	}

	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveTargetDirectory([]string{filePath}); err == nil {
		t.Fatal("expected an error for a file path")
	}
}

func TestValidateOrganizationDirectory(t *testing.T) {
	cleanDir := t.TempDir()
	if err := validateOrganizationDirectory(cleanDir); err != nil {
		t.Fatalf("clean directory rejected: %v", err)
	}

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module example"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateOrganizationDirectory(projectDir); err == nil {
		t.Fatal("expected project directory to be rejected")
	}
}
