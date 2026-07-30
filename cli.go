package main

import "fmt"

type outputFormat string

const (
	textFormat outputFormat = "text"
	jsonFormat outputFormat = "json"
)

func directoryOptions(args []string) (string, outputFormat, error) {
	format := textFormat
	var directory string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dir", "-d":
			i++
			if i == len(args) {
				return "", "", fmt.Errorf("%s requires a directory", args[i-1])
			}
			directory = args[i]
		case "--format":
			i++
			if i == len(args) {
				return "", "", fmt.Errorf("--format requires text or json")
			}
			format = outputFormat(args[i])
		default:
			if len(args[i]) > 0 && args[i][0] == '-' {
				return "", "", fmt.Errorf("unknown option %q", args[i])
			}
			if directory != "" {
				return "", "", fmt.Errorf("expected only one directory")
			}
			directory = args[i]
		}
	}
	if format != textFormat && format != jsonFormat {
		return "", "", fmt.Errorf("unsupported format %q (use text or json)", format)
	}
	if directory == "" {
		resolved, err := resolveTargetDirectory(nil)
		return resolved, format, err
	}
	resolved, err := resolveTargetDirectory([]string{directory})
	return resolved, format, err
}

func mutationOptions(args []string, allowDirectory bool) (string, bool, bool, []string, error) {
	apply, yes := false, false
	var directory string
	var names []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--apply":
			apply = true
		case "--yes":
			yes = true
		case "--dir", "-d":
			if !allowDirectory {
				return "", false, false, nil, fmt.Errorf("--dir is not supported")
			}
			i++
			if i == len(args) {
				return "", false, false, nil, fmt.Errorf("--dir requires a directory")
			}
			directory = args[i]
		default:
			if len(args[i]) > 0 && args[i][0] == '-' {
				return "", false, false, nil, fmt.Errorf("unknown option %q", args[i])
			}
			names = append(names, args[i])
		}
	}
	if directory == "" && allowDirectory {
		var err error
		directory, err = resolveTargetDirectory(nil)
		if err != nil {
			return "", false, false, nil, err
		}
	} else if allowDirectory {
		var err error
		directory, err = resolveTargetDirectory([]string{directory})
		if err != nil {
			return "", false, false, nil, err
		}
	}
	return directory, apply, yes, names, nil
}
