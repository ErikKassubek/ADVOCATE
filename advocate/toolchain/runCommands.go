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
	"advocate/utils/control"
	"advocate/utils/flags"
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

// runCommand runs a command line (shell) commands
//
// Parameter:
//   - osOut *os.File: file/output to write to not being what os.Stdout points to
//   - osErr *os.File: file/output to write to not being what os.Stdout points to
//   - name string: main command
//   - args ...string: command line parameters
//
// Returns:
//   - error
func runCommand(osOut, osErr *os.File, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(flags.TimeoutReplay)*time.Second)
	id := control.AddRunningCom(cancel)
	defer control.RemoveRunningCom(id)

	cmd := exec.CommandContext(ctx, name, args...)

	if flags.Output {
		if osOut != nil {
			multiOut := io.MultiWriter(os.Stdout, osOut)
			cmd.Stdout = multiOut
		}
		if osErr != nil {
			multiErr := io.MultiWriter(os.Stderr, osErr)
			cmd.Stderr = multiErr
		}
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// func runCommandWithOutput(name, outputFile string, args ...string) (string, error) {
// 	cmd := exec.Command(name, args...)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return "", err
// 	}

// 	// Write output to the specified file
// 	return string(output), os.WriteFile(outputFile, output, 0644)
// }

// // runCommandWithTee runs a command and writes output to a file
// func runCommandWithTee(name, outputFile string, args ...string) error {
// 	cmd := exec.Command(name, args...)
// 	outfile, err := os.Create(outputFile)
// 	if err != nil {
// 		return err
// 	}
// 	defer outfile.Close()
// 	cmd.Stdout = outfile
// 	cmd.Stderr = outfile
// 	return cmd.Run()
// }
