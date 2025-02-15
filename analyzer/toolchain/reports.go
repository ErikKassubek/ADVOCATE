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
	"analyzer/stats"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

/*
 * Generate the bug reports
 * Args:
 *    folderName string: path to folder containing the results
 */
func generateBugReports(folder string) {
	err := explanation.CreateOverview(folder, true)
	if err != nil {
		log.Println("Error creating explanation: ", err.Error())
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
 *    packagePath (string): path to the package directory
 *    destination (string): path to the destination directory
 *    total (bool): merge all already created logs into total log, for fuzzing
 */
func collect(packagePath, destination string, total bool) {
	filesToMove := []string{
		"advocateTrace",
		"results_machine.log",
		"results_readable.log",
		"output.log",
	}

	pattersToMove := []string{
		"rewritten_trace*",
		"advocateTraceReplay_*",
		"results_machine_*",
		"results_readable_*",
	}

	logsToCollect := []string{
		"results_machine.log",
		"results_readable.log",
		"output.log",
	}

	if total {
		for _, file := range logsToCollect {
			src := filepath.Join(packagePath, file)
			dest := filepath.Join(destination, "total_"+file)

			_, err := os.Stat(dest)
			new := os.IsNotExist(err)

			srcFile, err := os.Open(src)
			if err != nil {
				log.Println("Could not open src file ", src, ": ", err.Error())
				continue
			}
			defer srcFile.Close()

			destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println("Could not open dest file ", dest, ": ", err.Error())
				continue
			}
			defer destFile.Close()

			if !new {
				_, err = destFile.WriteString("==================================\n")
			}

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				log.Println("Could not merge ", src, " int ", dest, ": ", err.Error())
			}
		}
	}

	for _, file := range filesToMove {
		src := filepath.Join(packagePath, file)
		dest := filepath.Join(destination, file)
		err := os.Rename(src, dest)
		if err != nil {
			panic(fmt.Sprintf("Could not rename file %s to %s: %s", src, dest, err.Error()))
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

func updateStatsFiles(pathToAnalyzer string, progName string, testName string, dir string) {
	err := stats.CreateStats(dir, progName, testName)
	if err != nil {
		log.Println("Could not create statistics: ", err.Error())
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
