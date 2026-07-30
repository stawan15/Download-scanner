package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func defaultDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, "Downloads"), nil
}

func resolveTargetDirectory(args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("expected zero or one directory, received %d", len(args))
	}

	directory := ""
	if len(args) == 0 {
		var err error
		directory, err = defaultDirectory()
		if err != nil {
			return "", err
		}
	} else {
		directory = args[0]
	}

	absDirectory, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absDirectory)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", absDirectory)
	}

	return absDirectory, nil
}
