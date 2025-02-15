// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: runFullWorkflowMain.go
// Brief: Function to run commands
//
// Author: Erik Kassubek, Mario Occhinegro
// Created: 2024-09-18
//
// License: BSD-3-Clause

package toolchain

import (
	"os"
	"os/exec"
)

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandWithOutput(name, outputFile string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Write output to the specified file
	return string(output), os.WriteFile(outputFile, output, 0644)
}

// runCommandWithTee runs a command and writes output to a file
func runCommandWithTee(name, outputFile string, args ...string) error {
	cmd := exec.Command(name, args...)
	outfile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	return cmd.Run()
}
