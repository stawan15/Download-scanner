package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type SecurityResult struct {
	Directory string    `json:"directory"`
	Findings  []Finding `json:"findings"`
}

func runSecurity(args []string) error {
	useClamAV := len(args) > 0 && args[0] == "av"
	if useClamAV {
		args = args[1:]
	}
	directory, format, err := directoryOptions(args)
	if err != nil {
		return err
	}
	if useClamAV {
		if format == jsonFormat {
			return fmt.Errorf("JSON output is not supported for ClamAV scans")
		}
		return scanWithClamAV(directory)
	}
	result, err := securityResult(directory)
	if err != nil {
		return err
	}
	if format == jsonFormat {
		return writeJSON(result)
	}
	if len(result.Findings) == 0 {
		fmt.Println("No files need review.")
		return nil
	}
	fmt.Println("Files to review:")
	for _, finding := range result.Findings {
		fmt.Printf("\n[%d/100] %s\n", finding.Score, finding.Name)
		for _, reason := range finding.Reasons {
			fmt.Printf("  - %s\n", reason)
		}
	}
	return nil
}

func securityResult(directory string) (SecurityResult, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return SecurityResult{}, err
	}
	result := SecurityResult{Directory: directory}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if finding, found := securityFinding(entry.Name()); found {
			result.Findings = append(result.Findings, finding)
		}
	}
	sort.Slice(result.Findings, func(i, j int) bool {
		if result.Findings[i].Score == result.Findings[j].Score {
			return result.Findings[i].Name < result.Findings[j].Name
		}
		return result.Findings[i].Score > result.Findings[j].Score
	})
	return result, nil
}

func securityFinding(name string) (Finding, bool) {
	if name == ".DS_Store" || name == ".localized" {
		return Finding{}, false
	}
	scriptExtensions := map[string]bool{".exe": true, ".bat": true, ".cmd": true, ".ps1": true, ".vbs": true, ".js": true, ".jar": true, ".sh": true, ".command": true}
	installerExtensions := map[string]bool{".dmg": true, ".pkg": true, ".deb": true, ".iso": true}
	decoyExtensions := map[string]bool{".pdf": true, ".doc": true, ".docx": true, ".jpg": true, ".jpeg": true, ".png": true, ".zip": true}
	lowerName := strings.ToLower(name)
	extension := filepath.Ext(lowerName)
	score := 0
	var reasons []string
	if strings.HasPrefix(name, ".") {
		score += 15
		reasons = append(reasons, "hidden file")
	}
	if scriptExtensions[extension] {
		score += 40
		reasons = append(reasons, "executable or script file")
	}
	if installerExtensions[extension] {
		score += 10
		reasons = append(reasons, "downloaded installer")
	}
	previousExtension := filepath.Ext(strings.TrimSuffix(lowerName, extension))
	if decoyExtensions[previousExtension] && scriptExtensions[extension] {
		score += 40
		reasons = append(reasons, "double extension may disguise file type")
	}
	if score == 0 {
		return Finding{}, false
	}
	return Finding{Name: name, Score: score, Reasons: reasons}, true
}

func scanWithClamAV(directory string) error {
	if _, err := exec.LookPath("clamscan"); err != nil {
		return fmt.Errorf("ClamAV is not installed; install clamscan and try again")
	}
	fmt.Printf("Scanning %s with ClamAV...\nThis scan only reports findings; it does not delete files.\n\n", directory)
	command := exec.Command("clamscan", "--recursive", "--infected", "--no-summary", directory)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err == nil {
		fmt.Println("\nClamAV found no known malware.")
		return nil
	}
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
		fmt.Println("\nClamAV found one or more suspicious files.")
		return nil
	}
	return fmt.Errorf("ClamAV could not complete the scan: %w", err)
}
