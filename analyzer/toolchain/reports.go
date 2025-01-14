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
	"fmt"
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
 */
func moveResults(packagePath, destination string) {
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

	for _, file := range filesToMove {
		src := filepath.Join(packagePath, file)
		dest := filepath.Join(destination, file)
		_ = os.Rename(src, dest)
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

	for _, pattern := range pattersToMove {
		files, _ := filepath.Glob(filepath.Join(path, pattern))
		for _, trace := range files {
			os.RemoveAll(trace)
		}
	}
}

func updateStatsFiles(pathToAnalyzer string, progName string, testName string, dir string) {
	// TODO (COMMAND): replace by direct call
	err := runCommand(pathToAnalyzer, "stats", "-trace", dir, "-prog", progName, "-test", testName)
	if err != nil {
		fmt.Println("Could not create statistics")
	}
}
