package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func runTrash(args []string) error {
	directory, apply, yes, names, err := mutationOptions(args, true)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return fmt.Errorf("provide one or more file names")
	}
	plans, skipped := trashPlans(directory, names)
	for _, message := range skipped {
		fmt.Println(message)
	}
	if len(plans) == 0 {
		fmt.Println("Nothing to move.")
		return nil
	}
	fmt.Println("Trash preview:")
	for _, plan := range plans {
		fmt.Printf("- %s\n", filepath.Base(plan.Source))
	}
	if !apply {
		fmt.Println("\nPreview only. Re-run with --apply --yes to move these files to Trash.")
		return nil
	}
	if !yes {
		return fmt.Errorf("trash requires --yes together with --apply")
	}
	for _, plan := range plans {
		if err := os.Rename(plan.Source, plan.Destination); err != nil {
			return fmt.Errorf("could not move %s: %w", filepath.Base(plan.Source), err)
		}
		fmt.Printf("Moved to Trash: %s\n", filepath.Base(plan.Source))
	}
	return nil
}

func trashPlans(directory string, fileNames []string) ([]MovePlan, []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, []string{fmt.Sprintf("Could not find home directory: %v", err)}
	}
	trashDir := filepath.Join(homeDir, ".Trash")
	var plans []MovePlan
	var skipped []string
	for _, fileName := range fileNames {
		if filepath.Base(fileName) != fileName {
			skipped = append(skipped, fmt.Sprintf("Skipping invalid file name: %s", fileName))
			continue
		}
		source := filepath.Join(directory, fileName)
		info, err := os.Lstat(source)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("Could not find %s: %v", fileName, err))
			continue
		}
		if info.IsDir() {
			skipped = append(skipped, fmt.Sprintf("Skipping directory: %s", fileName))
			continue
		}
		destination := filepath.Join(trashDir, fileName)
		if _, err := os.Lstat(destination); err == nil {
			skipped = append(skipped, fmt.Sprintf("Skipping %s: a file with this name is already in Trash", fileName))
			continue
		}
		plans = append(plans, MovePlan{Source: source, Destination: destination})
	}
	return plans, skipped
}
