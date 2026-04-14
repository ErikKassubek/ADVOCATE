package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// filter is a function that determines whether a line should be copied.
func filter(elem string) bool {
	if elem == "" {
		return false
	}

	if !strings.Contains(elem, ",") {
		return false
	}

	elemSplit := strings.Split(elem, ",")
	elem = elemSplit[len(elemSplit)-1]

	if !strings.Contains(elem, ":") {
		return false
	}

	elemSplit = strings.Split(elem, ":")
	if len(elemSplit) != 2 {
		return false
	}

	file := elemSplit[0]

	if strings.Contains(file, "go-patch/src/") {
		return false
	} else if strings.Contains(file, "go/pkg/mod/") {
		return false
	} else if strings.Contains(file, "time/sleep.go") {
		return false
	} else if strings.Contains(file, "signal/signal.go") { // ctrl+c
		return false
	}

	if strings.HasSuffix(file, "advocate/advocate.go") ||
		strings.HasSuffix(file, "runtime/advocate_replay.go") ||
		strings.HasSuffix(file, "runtime/advocate_routine.go") ||
		strings.HasSuffix(file, "runtime/advocate_trace.go") ||
		strings.HasSuffix(file, "runtime/advocate_utile.go") ||
		strings.HasSuffix(file, "runtime/advocate_atomic.go") { // internal
		return false
	} else if strings.HasSuffix(file, "syscall/env_unix.go") {
		return false
	} else if strings.HasSuffix(file, "runtime/signal_unix.go") {
		return false
	} else if strings.HasSuffix(file, "runtime/mgc.go") { // garbage collector
		return false
	} else if strings.HasSuffix(file, "runtime/panic.go") {
		return false
	}

	return true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: filter_copy <folder_path>")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	// Ensure the input path exists and is a directory
	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Println("Error: Provided path is not a directory.")
		os.Exit(1)
	}

	// Create the parallel output folder
	outputPath := inputPath + "_filtered"
	err = os.Mkdir(outputPath, 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Traverse files in the input folder
	err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip directories, but replicate folder structure
			relPath, err := filepath.Rel(inputPath, path)
			if err != nil {
				return err
			}
			newDir := filepath.Join(outputPath, relPath)
			return os.MkdirAll(newDir, 0755)
		}

		// Process files
		return processFile(path, inputPath, outputPath)
	})

	if err != nil {
		fmt.Printf("Error during traversal: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Filtered files have been copied to: %s\n", outputPath)
}

// processFile reads a file, filters its lines, and writes the result to the output folder.
func processFile(filePath, inputPath, outputPath string) error {
	relPath, err := filepath.Rel(inputPath, filePath)
	if err != nil {
		return err
	}

	// Compute the destination file path
	destPath := filepath.Join(outputPath, relPath)

	if !strings.Contains(relPath, "trace_") {
		return nil
	}

	// Open the input file
	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open input file %s: %w", filePath, err)
	}
	defer inputFile.Close()

	// Create the output file
	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", destPath, err)
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(outputFile)

	// Filter lines
	written := false
	for scanner.Scan() {
		line := scanner.Text()
		if filter(line) {
			written = true
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				return fmt.Errorf("error writing to file %s: %w", destPath, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from file %s: %w", filePath, err)
	}

	// Flush writer to ensure all content is written
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing to file %s: %w", destPath, err)
	}

	if !written {
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("error deleting empty file %s: %w", destPath, err)
		}
	}

	return nil
}
