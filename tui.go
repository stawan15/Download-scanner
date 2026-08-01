package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const tuiAccent = "\033[36m"
const tuiMuted = "\033[2m"
const tuiReset = "\033[0m"

// runTUI provides an interactive front end while keeping the CLI commands
// available for scripting. Destructive operations still require typed consent.
func runTUI(in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)
	directory, err := defaultDirectory()
	if err != nil {
		return err
	}

	for {
		printTUIMenu(out, directory)
		choice, err := readTUIInput(reader, out, "Choose an option")
		if err != nil {
			return err
		}

		switch strings.ToLower(choice) {
		case "1", "scan":
			err = scanDirectory(directory, textFormat)
		case "2", "duplicates":
			err = findDuplicates(directory, textFormat)
		case "3", "security":
			err = runSecurity([]string{"--dir", directory})
		case "4", "organize":
			err = planOrganization(directory)
		case "5", "apply":
			confirmed, confirmErr := confirmTUIAction(reader, out, "Type MOVE to organize the previewed files", "MOVE")
			if confirmErr != nil {
				return confirmErr
			}
			if confirmed {
				err = applyOrganization(directory)
			} else {
				fmt.Fprintln(out, "Nothing was changed.")
			}
		case "6", "undo":
			confirmed, confirmErr := confirmTUIAction(reader, out, "Type UNDO to reverse the latest organization", "UNDO")
			if confirmErr != nil {
				return confirmErr
			}
			if confirmed {
				err = undoOrganization()
			} else {
				fmt.Fprintln(out, "Nothing was changed.")
			}
		case "7", "folder", "directory":
			value, inputErr := readTUIInput(reader, out, "Folder path (leave blank to keep current)")
			if inputErr != nil {
				return inputErr
			}
			if value != "" {
				directory, err = resolveTargetDirectory([]string{value})
			}
		case "q", "quit", "exit":
			fmt.Fprintln(out, "Goodbye. Your files were not changed.")
			return nil
		default:
			fmt.Fprintln(out, "Please choose a listed option.")
		}

		if err != nil {
			fmt.Fprintf(out, "\nError: %v\n", err)
		}
		if _, err := readTUIInput(reader, out, "Press Enter to return to the menu"); err != nil {
			return err
		}
	}
}

func printTUIMenu(out io.Writer, directory string) {
	fmt.Fprintf(out, "\n%sDownload Inbox Cleaner%s\n", tuiAccent, tuiReset)
	fmt.Fprintf(out, "%sFolder: %s%s\n\n", tuiMuted, directory, tuiReset)
	fmt.Fprintln(out, "  1. Scan files")
	fmt.Fprintln(out, "  2. Find duplicates")
	fmt.Fprintln(out, "  3. Check security risks")
	fmt.Fprintln(out, "  4. Preview organization")
	fmt.Fprintln(out, "  5. Organize files  (requires MOVE confirmation)")
	fmt.Fprintln(out, "  6. Undo last organization  (requires UNDO confirmation)")
	fmt.Fprintln(out, "  7. Change folder")
	fmt.Fprintln(out, "  q. Quit")
}

func readTUIInput(reader *bufio.Reader, out io.Writer, prompt string) (string, error) {
	fmt.Fprintf(out, "\n%s: ", prompt)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	if err == io.EOF && value == "" {
		return "", io.EOF
	}
	return strings.TrimSpace(value), nil
}

func confirmTUIAction(reader *bufio.Reader, out io.Writer, prompt, expected string) (bool, error) {
	value, err := readTUIInput(reader, out, prompt)
	if err != nil {
		return false, err
	}
	return value == expected, nil
}
