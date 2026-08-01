package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "help" {
		printUsage()
		return nil
	}
	switch args[0] {
	case "version", "--version":
		fmt.Println(version)
		return nil
	case "tui", "--tui":
		return runTUI(os.Stdin, os.Stdout)
	case "scan":
		directory, format, err := directoryOptions(args[1:])
		if err != nil {
			return err
		}
		return scanDirectory(directory, format)
	case "duplicates":
		directory, format, err := directoryOptions(args[1:])
		if err != nil {
			return err
		}
		return findDuplicates(directory, format)
	case "security":
		return runSecurity(args[1:])
	case "organize":
		return runOrganization(args[1:])
	case "trash":
		return runTrash(args[1:])
	case "undo":
		return runUndo(args[1:])
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println(`Download Inbox Cleaner

Usage:
  cleaner scan [--dir DIRECTORY] [--format text|json]
  cleaner duplicates [--dir DIRECTORY] [--format text|json]
  cleaner security [av] [--dir DIRECTORY] [--format text|json]
  cleaner organize [--apply --yes] [--dir DIRECTORY]
  cleaner trash [--apply --yes] [--dir DIRECTORY] <file-name> [file-name...]
  cleaner undo [--yes]
  cleaner tui
  cleaner version

Directories default to ~/Downloads. Commands that change files require both --apply and --yes.`)
}
