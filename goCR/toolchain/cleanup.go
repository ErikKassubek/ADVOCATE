//
// File: cleanup.go
// Brief: Cleanup traces and files
//
// Created: 2025-02-28
//
// License: BSD-3-Clause

package toolchain

import (
	"goCR/utils/log"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// Function to move results files from the package directory to the destination directory
//
// Parameter:
//   - progPath string: path to the program
//   - packagePath string: path to the package directory
//   - destination string: path to the destination directory
//   - total bool: merge all already created logs into total log, for fuzzing
func collect(progPath, packagePath, destination string, total bool) {
	filesToMove := []string{
		"goCRTrace",
		"results_machine.log",
		"results_readable.log",
		"output.log",
	}

	pattersToMove := []string{
		"rewrittenTrace*",
	}

	logsToCollect := []string{
		"results_machine.log",
		"results_readable.log",
		"output.log",
	}

	if total {
		for _, file := range logsToCollect {
			src := filepath.Join(progPath, file)
			dest := filepath.Join(destination, "total_"+file)

			_, err := os.Stat(dest)
			new := os.IsNotExist(err)

			srcFile, err := os.Open(src)
			if err != nil {
				continue
			}
			defer srcFile.Close()

			destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				continue
			}
			defer destFile.Close()

			if !new {
				_, err = destFile.WriteString("==================================\n")
			}

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				log.Error("Could not merge ", src, " int ", dest, ": ", err.Error())
			}
		}
	}

	for _, file := range filesToMove {
		src := filepath.Join(packagePath, file)
		if file == "output.log" {
			src = filepath.Join(progPath, file)
		}
		dest := filepath.Join(destination, file)

		if file == "goCRTrace" {
			movedTraces++
			dest += "_" + strconv.Itoa(movedTraces)
		}

		err := os.Rename(src, dest)
		if err != nil {
			continue
		}
	}

	for _, pattern := range pattersToMove {
		files, _ := filepath.Glob(filepath.Join(packagePath, pattern))
		for _, trace := range files {
			dest := filepath.Join(destination, filepath.Base(trace))
			_ = os.Rename(trace, dest)
		}
	}
}

// RemoveTraces removes all traces, both recorded and rewritten from the path
//
// Parameter:
//   - path string: path to the folder containing the traces
//   - total bool: if true, remove all traces, otherwise only remove traces that are only needed for one fuzzing run
func RemoveTraces(path string, total bool) {
	pattersToMove := []string{
		"goCRTrace_*",
		"rewrittenTrace*",
		"fuzzingData.log",
	}

	if total {
		pattersToMove = append(pattersToMove, "fuzzingTrace_*")
	}

	files := make([]string, 0)
	filepath.WalkDir(path, func(p string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Use Glob to check if the file/directory matches the pattern
		for _, pattern := range pattersToMove {
			match, err := filepath.Match(pattern, filepath.Base(p))
			if err != nil {
				return err
			}

			if match {
				files = append(files, p)
			}
		}

		return nil
	})

	for _, trace := range files {
		os.RemoveAll(trace)
	}
}

// removeLogs removes the result and output files
//
// Parameter:
//   - path to the folder containing the result and output files
func removeLogs(path string) {
	logsToRemove := []string{
		"results_machine.log",
		"results_readable.log",
		"output.log",
	}

	for _, logFile := range logsToRemove {
		os.Remove(filepath.Join(path, logFile))
	}
}

// ClearFuzzingTrace removes the fuzzing trace folder
//
// Parameter:
//   - path string: path to the folder containing the fuzzing traces
//   - keepTrace bool: if true move fuzzingTraces into the result folder, otherwise remove it
func ClearFuzzingTrace(path string, keepTrace bool) {
	fuzzingPath := filepath.Join(path, "fuzzingTraces")

	if keepTrace {
		_ = os.Rename(fuzzingPath, filepath.Join(currentResFolder, "fuzzingTraces"))
		// if err != nil {
		// 	log.Errorf("failed to move folder %s to %s: %s", fuzzingPath, fuzzingPath, err.Error())
		// }
	} else {
		err := os.RemoveAll(fuzzingPath)
		if err != nil {
			log.Errorf("Could not delete fuzzingTraces: %s", err.Error())
		}
	}
}
