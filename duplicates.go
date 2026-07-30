package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type DuplicateResult struct {
	Directory        string   `json:"directory"`
	Groups           [][]File `json:"groups"`
	ReclaimableBytes int64    `json:"reclaimable_bytes"`
	Skipped          []string `json:"skipped,omitempty"`
}

func findDuplicates(directory string, format outputFormat) error {
	result, err := duplicateResult(directory)
	if err != nil {
		return err
	}
	if format == jsonFormat {
		return writeJSON(result)
	}
	if len(result.Groups) == 0 {
		fmt.Println("No exact duplicate files found.")
		return nil
	}
	fmt.Println("Duplicate files:")
	for i, files := range result.Groups {
		fmt.Printf("\nGroup %d - %.2f MB each\n", i+1, float64(files[0].Size)/1024/1024)
		for _, file := range files {
			fmt.Printf("  %s\n", file.Name)
		}
	}
	fmt.Printf("\nFound %d duplicate groups. Reclaimable space: %.2f MB\n", len(result.Groups), float64(result.ReclaimableBytes)/1024/1024)
	return nil
}

func duplicateResult(directory string) (DuplicateResult, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return DuplicateResult{}, err
	}
	byHash := map[string][]File{}
	result := DuplicateResult{Directory: directory}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		info, err := entry.Info()
		if err != nil {
			result.Skipped = append(result.Skipped, entry.Name())
			continue
		}
		hash, err := hashFile(path)
		if err != nil {
			result.Skipped = append(result.Skipped, entry.Name())
			continue
		}
		byHash[hash] = append(byHash[hash], File{Name: entry.Name(), Size: info.Size()})
	}
	for _, files := range byHash {
		if len(files) < 2 {
			continue
		}
		sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
		result.Groups = append(result.Groups, files)
		result.ReclaimableBytes += files[0].Size * int64(len(files)-1)
	}
	sort.Slice(result.Groups, func(i, j int) bool { return result.Groups[i][0].Name < result.Groups[j][0].Name })
	sort.Strings(result.Skipped)
	return result, nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
