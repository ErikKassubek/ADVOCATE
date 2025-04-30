// Copyright (c) 2025 Erik Kassubek
//
// File: paths.go
// Brief: Utils using paths
//
// Author: Erik Kassubek
// Created: 2025-04-23
//
// License: BSD-3-Clause

package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// MakePathLocal transforms a path into a local path by adding a ./ at the beginning it has non
//
// Parameter:
//   - path string: path
//
// Returns:
//   - string: path starting with ./
func MakePathLocal(path string) string {
	pathSep := string(os.PathSeparator)

	// ./path
	if strings.HasPrefix(path, "."+pathSep) {
		return path
	}

	// /path
	if strings.HasPrefix(path, pathSep) {
		return "." + path
	}

	// path
	return "." + pathSep + path
}

// GetDirectory returns the folder a file is in from the path
//
// Parameter:
//   - path string: the path to the file
//
// Returns:
//   - string: if path points to file, the folder it is in, if it points to a folder, the path
func GetDirectory(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return path
	}

	if info.IsDir() {
		// Already a directory
		return filepath.Clean(path)
	}

	// It's a file, return its directory
	return filepath.Dir(path)
}

// GetMainPath takes a path. If the path points to a file, it will return the path.
// If not it will check if the folder it points to contains a main.go file.
// If it does, it will return the path to the file
//
// Parameter:
//   - path string: path
//
// Returns:
//   - string: path to the main file
//   - error
func GetMainPath(path string) (string, error) {
	path = CleanPathHome(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		mainPath := filepath.Join(path, "main.go")

		if _, err := os.Stat(mainPath); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("main.go not found in directory %s", path)
			} else {
				return "", err
			}
		}
	}

	// It's a file, return the path as is
	return filepath.Clean(path), nil
}

// CleanPathHome takes a path containing a ~ and replaces it with the
// path to the home folder
func CleanPathHome(path string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(path, "~", home, -1)
}

// CheckPath checks if the provided path to the program that should
// be run/analyzed exists. If not, it panics.
//
// Parameter:
//   - path string: path to check
//
// Returns:
//   - string: the cleaned path
//   - error: error if path not exists, nil otherwise
func CheckPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("Path cannot be empty")
	}

	progPath := CleanPathHome(path)

	_, err := os.Stat(progPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return progPath, fmt.Errorf("Path %s does not exists", progPath)
		} else {
			return progPath, err
		}
	}

	return progPath, nil
}
