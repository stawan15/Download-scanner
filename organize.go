package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func runOrganization(args []string) error {
	// Keep the original `organize apply [directory]` spelling usable while the
	// explicit --yes safeguard protects the actual move operation.
	if len(args) > 0 && args[0] == "apply" {
		args[0] = "--apply"
	}
	directory, apply, yes, names, err := mutationOptions(args, true)
	if err != nil {
		return err
	}
	if len(names) > 0 {
		return fmt.Errorf("organize does not accept file names")
	}
	if !apply {
		return planOrganization(directory)
	}
	if !yes {
		return fmt.Errorf("organize requires --yes together with --apply")
	}
	return applyOrganization(directory)
}

func planOrganization(directory string) error {
	plans, err := organizationPlans(directory)
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		fmt.Println("No files match the organization rules.")
		return nil
	}
	fmt.Println("Organization preview:")
	for _, plan := range plans {
		relative, _ := filepath.Rel(directory, filepath.Dir(plan.Destination))
		fmt.Printf("- %s -> %s\n", filepath.Base(plan.Source), relative)
	}
	fmt.Printf("\n%d files would be organized. Preview only; nothing was changed.\n", len(plans))
	return nil
}

func organizationPlans(directory string) ([]MovePlan, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var plans []MovePlan
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		folder := organizationFolder(entry.Name())
		if folder == "" {
			continue
		}
		plans = append(plans, MovePlan{Source: filepath.Join(directory, entry.Name()), Destination: filepath.Join(directory, folder, entry.Name())})
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].Destination < plans[j].Destination })
	return plans, nil
}

func applyOrganization(directory string) error {
	return applyOrganizationTo(directory, os.Stdout)
}

func applyOrganizationTo(directory string, out io.Writer) error {
	if err := validateOrganizationDirectory(directory); err != nil {
		return err
	}
	plans, err := organizationPlans(directory)
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		fmt.Fprintln(out, "No files match the organization rules.")
		return nil
	}
	var valid []MovePlan
	for _, plan := range plans {
		if _, err := os.Lstat(plan.Source); err != nil {
			fmt.Fprintf(out, "Skipping missing file: %s\n", filepath.Base(plan.Source))
			continue
		}
		if _, err := os.Lstat(plan.Destination); err == nil {
			fmt.Fprintf(out, "Skipping existing destination: %s\n", plan.Destination)
			continue
		}
		valid = append(valid, plan)
	}
	if len(valid) == 0 {
		fmt.Fprintln(out, "Nothing can be moved safely.")
		return nil
	}
	var completed []MovePlan
	var failures []string
	for _, plan := range valid {
		if err := os.MkdirAll(filepath.Dir(plan.Destination), 0755); err != nil {
			failures = append(failures, err.Error())
			continue
		}
		if _, err := os.Lstat(plan.Destination); err == nil {
			failures = append(failures, "destination appeared: "+plan.Destination)
			continue
		}
		if err := os.Rename(plan.Source, plan.Destination); err != nil {
			failures = append(failures, err.Error())
			continue
		}
		completed = append(completed, plan)
		fmt.Fprintf(out, "Moved: %s\n", filepath.Base(plan.Source))
	}
	if len(completed) > 0 {
		if err := writeOrganizationLog(completed); err != nil {
			return fmt.Errorf("files were moved, but undo log could not be saved: %w", err)
		}
		fmt.Fprintf(out, "\nMoved %d files. Run 'cleaner undo --yes' to reverse this organization.\n", len(completed))
	}
	if len(failures) > 0 {
		return fmt.Errorf("%d files could not be organized", len(failures))
	}
	return nil
}

func validateOrganizationDirectory(directory string) error {
	for _, marker := range []string{".git", "go.mod", "package.json", "Cargo.toml"} {
		if _, err := os.Stat(filepath.Join(directory, marker)); err == nil {
			return fmt.Errorf("%s looks like a project directory because it contains %s", directory, marker)
		}
	}
	return nil
}

func runUndo(args []string) error {
	_, apply, yes, names, err := mutationOptions(args, false)
	if err != nil {
		return err
	}
	if len(names) > 0 || apply {
		return fmt.Errorf("usage: cleaner undo --yes")
	}
	if !yes {
		return fmt.Errorf("undo requires --yes")
	}
	return undoOrganization()
}

func undoOrganization() error {
	return undoOrganizationTo(os.Stdout)
}

func undoOrganizationTo(out io.Writer) error {
	logPath, err := organizationLogPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(out, "There is no organization run to undo.")
			return nil
		}
		return err
	}
	var log OrganizationLog
	if err := json.Unmarshal(data, &log); err != nil {
		return err
	}
	if len(log.Moves) == 0 {
		fmt.Fprintln(out, "There is no organization run to undo.")
		return nil
	}
	var remaining []MovePlan
	for i := len(log.Moves) - 1; i >= 0; i-- {
		plan := log.Moves[i]
		if _, err := os.Lstat(plan.Source); err == nil {
			remaining = append(remaining, plan)
			continue
		}
		if _, err := os.Lstat(plan.Destination); err != nil {
			remaining = append(remaining, plan)
			continue
		}
		if err := os.Rename(plan.Destination, plan.Source); err != nil {
			remaining = append(remaining, plan)
			continue
		}
		fmt.Fprintf(out, "Restored: %s\n", filepath.Base(plan.Source))
	}
	if len(remaining) == 0 {
		if err := os.Remove(logPath); err != nil {
			return err
		}
		fmt.Fprintln(out, "\nOrganization undone successfully.")
		return nil
	}
	if err := writeOrganizationLog(remaining); err != nil {
		return err
	}
	return fmt.Errorf("%d files could not be restored; run undo again after resolving them", len(remaining))
}

func organizationLogPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "share", "download-cleaner", "last-organization.json"), nil
}
func writeOrganizationLog(moves []MovePlan) error {
	path, err := organizationLogPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(OrganizationLog{Moves: moves}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
func organizationFolder(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".pdf":
		return "Documents/PDF"
	case ".doc", ".docx", ".xlsx":
		return "Documents/Office"
	case ".png", ".jpg", ".jpeg", ".gif":
		return "Images"
	case ".zip", ".gz", ".tar", ".rar":
		return "Archives"
	case ".dmg", ".pkg", ".deb", ".iso":
		return "Installers"
	case ".ttf", ".otf", ".woff", ".woff2":
		return "Fonts"
	case ".excalidraw", ".svg":
		return "Design"
	case ".go", ".md":
		return "Code"
	default:
		return ""
	}
}
