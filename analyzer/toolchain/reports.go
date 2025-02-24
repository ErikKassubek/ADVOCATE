// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: headerUnitTests.go
// Brief: Functions to generate bug reports
//
// Author: Erik Kassubek, Mario Occhinegro
// Created: 2024-09-18
//
// License: BSD-3-Clause

package toolchain

import (
	"analyzer/explanation"
	"analyzer/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
)

/*
 * Generate the bug reports
 * Args:
 *    folderName string: path to folder containing the results
 *    fuzzingRun int: number of fuzzing run, -1 for not fuzzing
 */
func generateBugReports(folder string, fuzzing int) {
	err := explanation.CreateOverview(folder, true, fuzzing)
	if err != nil {
		utils.LogError("Error creating explanation: ", err.Error())
	}
}

/*
 * Get all files in folder path with name fileName
 * Args:
 *    folderPath (string): path to the folder to search in
 *    fileName (string): name of the files to search for
 * Returns:
 *    []string: list of the paths of the files
 *    error
 */
// func getFiles(folderPath string, fileName string) ([]string, error) {
// 	var files []string
// 	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if info.IsDir() {
// 			return nil
// 		}
// 		if filepath.Base(path) == fileName {
// 			files = append(files, path)
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return files, nil
// }

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
	patterns := []string{
		"advocateTrace",
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
		for _, pattern := range patterns {
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
