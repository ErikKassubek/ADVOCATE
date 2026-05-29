// Copyright (c) 2024 Erik Kassubek
//
// File: modes.go
// Brief: controll the different modes
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package run

import (
	"advocate/fuzzing"
	"advocate/results/stats"
	"advocate/toolchain"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
)

// modeFuzzing starts the fuzzing
func modeFuzzing() error {
	if flags.ProgName == "" {
		flags.ProgName = paths.GetProgName(flags.ProgPath)
	}

	var err error
	flags.ProgPath, err = paths.CheckPath(flags.ProgPath)
	if err != nil {
		log.Error("Error on checking prog path: ", err)
		log.Error("Set path with -path [path]")
		return err
		panic(err)
	}

	err = fuzzing.Fuzzing()
	if err != nil {
		log.Error("Fuzzing Failed: ", err.Error())
		return err
	}

	return nil
}

// Start point for the toolchain
// This will run, analyze and replay a given program or test
//
// Parameter:
//   - mode string: main for main function, test for test function
//   - record bool: if true, the toolchain will run the recording
//   - analysis bool: if true, the toolchain will run analysis
//   - replay bool: if true, the toolchain will run replays
//
// Note:
//   - If recording is false, but analysis or replay is set, -trace must be set
func modeToolchain(mode string, record bool, analysis bool, replay bool) error {
	var err error
	flags.ProgPath, err = paths.CheckPath(flags.ProgPath)
	if err != nil {
		log.Error("Error on checking prog path: ", err)
		return err
	}

	if !record && (analysis || replay) {
		flags.TracePath, err = paths.CheckPath(flags.TracePath)
		if err != nil {
			log.Error("Error on checking trace path: ", err)
			return err
		}
	}

	if mode == "test" && !record && replay && flags.ExecName == "" {
		log.Error("When running replay of test without recording, -exec [TestName] must be set")
		panic("When running replay of test without recording, -exec [TestName] must be set")
	}

	firstRun := true
	fileNumber, testNumber := 1, 0
	_, _, err = toolchain.Run(mode, "", record, analysis,
		replay, -1, "", firstRun, fileNumber, testNumber)
	if err != nil {
		return err
	}

	if flags.CreateStatistics {
		err = stats.CreateStatsTotal(flags.ProgPath)
		if err != nil {
			return err
		}
	}

	return nil
}
