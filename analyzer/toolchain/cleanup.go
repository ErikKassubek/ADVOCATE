// Copyright (c) 2025 Erik Kassubek
//
// File: cleanup.go
// Brief: Cleanup traces and files
//
// Author: Erik Kassubek
// Created: 2025-02-28
//
// License: BSD-3-Clause

package toolchain

import (
	"analyzer/utils"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
 * Function to move results files from the package directory to the destination directory
 * Args:
 *    progPath (string): path to the program
 *    packagePath (string): path to the package directory
 *    destination (string): path to the destination directory
 *    total (bool): merge all already created logs into total log, for fuzzing
 */
func collect(progPath, packagePath, destination string, total bool) {
	filesToMove := []string{
		"advocateTrace",
		"results_machine.log",
		"results_readable.log",
		"output.log",
	}

	pattersToMove := []string{
		"rewritten_trace*",
		"advocateTraceReplay_*",
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
				utils.LogError("Could not open src file ", src, ": ", err.Error())
				continue
			}
			defer srcFile.Close()

			destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				utils.LogError("Could not open dest file ", dest, ": ", err.Error())
				continue
			}
			defer destFile.Close()

			if !new {
				_, err = destFile.WriteString("==================================\n")
			}

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				utils.LogError("Could not merge ", src, " int ", dest, ": ", err.Error())
			}
		}
	}

	for _, file := range filesToMove {
		src := filepath.Join(packagePath, file)
		if strings.HasSuffix(file, ".log") {
			src = filepath.Join(progPath, file)
		}
		dest := filepath.Join(destination, file)

		if file == "advocateTrace" {
			movedTraces += 1
			dest += "_" + strconv.Itoa(movedTraces)
		}

		err := os.Rename(src, dest)
		if err != nil {
			utils.LogErrorf("Could not rename file %s to %s: %s\n", src, dest, err.Error())
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

/*
 * Remove all traces, both recorded and rewritten from the path
 * Args:
 * 	path (string): path to the folder containing the traces
 */
func removeTraces(path string) {
	pattersToMove := []string{
		"advocateTrace_*",
		"rewritten_trace*",
		"advocateTraceReplay_*",
		"fuzzingData.log",
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

func removeLogs(path string) {
	logsToRemove := []string{
		"results_machine.log",
		"results_readable.log",
		"output.log",
	}

	for _, logFile := range logsToRemove {
		_ = os.Remove(filepath.Join(path, logFile))
	}
}
