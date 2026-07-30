package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ScanResult struct {
	Directory   string           `json:"directory"`
	ItemCount   int              `json:"item_count"`
	TotalBytes  int64            `json:"total_bytes"`
	Files       []File           `json:"files"`
	Directories []string         `json:"directories"`
	Extensions  []ExtensionCount `json:"extensions"`
}
type ExtensionCount struct {
	Extension string `json:"extension"`
	Count     int    `json:"count"`
}

func scanDirectory(directory string, format outputFormat) error {
	result, err := scanResult(directory)
	if err != nil {
		return err
	}
	if format == jsonFormat {
		return writeJSON(result)
	}
	fmt.Printf("Scanning %s\n", result.Directory)
	for _, name := range result.Directories {
		fmt.Printf("[DIR] %s\n", name)
	}
	fmt.Println("\nFiles by size:")
	for _, file := range result.Files {
		fmt.Printf("[FILE] %8.2f MB %s\n", float64(file.Size)/1024/1024, file.Name)
	}
	fmt.Printf("\nTotal file size: %.2f MB\n\nFiles by extension:\n", float64(result.TotalBytes)/1024/1024)
	for _, extension := range result.Extensions {
		fmt.Printf("%-15s %d files\n", extension.Extension, extension.Count)
	}
	fmt.Printf("\nFound %d items\n", result.ItemCount)
	return nil
}

func scanResult(directory string) (ScanResult, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return ScanResult{}, err
	}
	result := ScanResult{Directory: directory, ItemCount: len(entries)}
	counts := map[string]int{}
	for _, entry := range entries {
		if entry.IsDir() {
			result.Directories = append(result.Directories, entry.Name())
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		extension := filepath.Ext(entry.Name())
		if extension == "" {
			extension = "[no extension]"
		}
		result.TotalBytes += info.Size()
		counts[extension]++
		result.Files = append(result.Files, File{Name: entry.Name(), Size: info.Size()})
	}
	sort.Strings(result.Directories)
	sort.Slice(result.Files, func(i, j int) bool {
		if result.Files[i].Size == result.Files[j].Size {
			return result.Files[i].Name < result.Files[j].Name
		}
		return result.Files[i].Size > result.Files[j].Size
	})
	for extension, count := range counts {
		result.Extensions = append(result.Extensions, ExtensionCount{extension, count})
	}
	sort.Slice(result.Extensions, func(i, j int) bool { return result.Extensions[i].Extension < result.Extensions[j].Extension })
	return result, nil
}

func writeJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
