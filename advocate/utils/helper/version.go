// Copyright (c) 2025 Erik Kassubek
//
// File: version.go
// Brief: Check the go version and exec name of a given program
//
// Author: Erik Kassubek
// Created: 2025-05-19
//
// License: BSD-3-Clause

package helper

import (
	"advocate/utils/flags"
	"advocate/utils/log"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckGoMod checks the version of the program to be analyzed and finds the exec name
// Advocate is implemented in and for go1.25. It the analyzed program has another
// version, especially if the other version is also installed on the machine,
// this can lead to problems. checkGoMod therefore reads the version of the
// analyzed program and if its not 1.25, a warning and information is printed
// to the terminal
// Additionally it reads the module name from the go.mod file.
// If -main is set, but -exec is not set it will try to set the
// execname value. If no module value is found, the program will panic
//
// Returns:
//   - string: exec name, or empty if not found
func CheckGoMod() string {
	var goModPath string

	if flags.ProgPath == "" {
		return flags.ExecName
	}

	// Search for go.mod
	err := filepath.WalkDir(GetDirectory(flags.ProgPath), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == "go.mod" {
			goModPath = path
			return filepath.SkipAll // Stop searching after finding the first one
		}
		return nil
	})

	if goModPath == "" {
		log.Info("Could not find go.mod")
		return flags.ExecName
	}

	// Open and read go.mod
	file, err := os.Open(goModPath)
	if err != nil {
		log.Info("Could not find go.mod")
		return flags.ExecName
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// check for module name
		if flags.ModeMain && flags.ExecName == "" && strings.HasPrefix(line, "module") {
			s := strings.Split(line, " ")
			if len(s) < 2 {
				continue
			}

			flags.ExecName = s[1]
			continue
		}

		// check for version
		if strings.HasPrefix(line, "go ") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "go "))

			versionSplit := strings.Split(version, ".")

			if len(versionSplit) < 2 {
				log.Error("Invalid go version")
			}

			if versionSplit[0] != "1" || versionSplit[1] != "25" {
				errString := "ADVOCATE is implemented for go version 1.25"
				errString += fmt.Sprintf("Found version %s. ", version)
				errString += fmt.Sprintf("This may result in the analysis not working correctly, especially if go %s.%s is installed on the computer. ", versionSplit[0], versionSplit[1])
				errString += "The message 'package advocate is not in std' in the output.log file may indicate this."
				// errString += `'/home/.../go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/advocate' or 'package advocate is not in std' in the output files may indicate an incompatible go version.`
				log.Important(errString)
			}

			return flags.ExecName
		}
	}

	log.Error("Could not determine go version")
	return flags.ExecName
}
