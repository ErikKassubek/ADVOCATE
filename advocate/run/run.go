// Copyright (c) 2026 Erik Kassubek
//
// File: run.go
// Brief: start/controll the different modes
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package run

import (
	"advocate/analysis/baseA"
	"advocate/fuzzing"
	"advocate/fuzzing/baseF"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/settings"
	"advocate/utils/timer"
	"fmt"
	"os"
)

// Run starts the execution of advocate
func Run() error {

	// If -main is set, the path needs to be the path to the main file
	// If the given path is to a folder, check if a main.go file exists in this folder
	// If so, fix the path. Otherwise return error and finish
	var err error
	if flags.ModeMain {
		flags.ProgPath, err = paths.GetMainPath(flags.ProgPath)
		if err != nil {
			log.Error("Could not find main file. If -main is set, -path should point to the main file.")
			log.Error(err)

			return err
		}
	}

	settings.SetSettings()
	paths.BuildPaths(flags.ModeMain)

	err = CheckBin()
	if err != nil {
		return err
	}

	progPathDir := paths.GetDirectory(flags.ProgPath)
	if err != nil {
		return err
	}
	timer.Init(progPathDir)
	timer.Start(timer.Total)
	defer timer.Stop(timer.Total)

	control.SetMaxNumberElem()
	if !flags.NoMemorySupervisor {
		go control.Supervisor(baseA.ClearTrace, baseA.ClearData, fuzzing.ResetFuzzing) // cancel analysis if not enough ram
	}

	// don't run any HB Analysis for direct GFuzz, GoPie and GoCR
	if flags.Mode == "fuzzing" && (flags.FuzzingMode == baseF.GFuzz ||
		flags.FuzzingMode == baseF.GoPie || flags.FuzzingMode == baseF.GoCR) {
		flags.Scenarios = "-"
		flags.OnlyAPanicAndLeak = true
	}

	if flags.FuzzingMode == "" {
		flags.FuzzingMode = baseF.Guided
	}

	baseA.AnalysisCasesMap, err = flags.ParseAnalysisCases()
	if err != nil {
		log.Error("Could not read analysis cases: ", err)
		return err
	}

	modeMainTest := "test"
	if flags.ModeMain {
		modeMainTest = "main"
	}

	CheckProg()
	helper.RunGoModTidy()

	if flags.ModeMain && flags.ExecName == "" {
		log.Error("Could not determine executable name from go.mod. Provide with -exec [ExecutableName]")
		return fmt.Errorf("Could not determine executable name")
	}

	switch flags.Mode {
	case "analysis":
		err = modeToolchain(modeMainTest, true, true, true)
	case "fuzzing":
		err = modeFuzzing()
	case "record", "recording":
		flags.DeleteTraces = false
		err = modeToolchain(modeMainTest, true, false, false)
	case "replay":
		err = modeToolchain(modeMainTest, false, false, true)
	default:
		log.Errorf("Unknown mode %s\n", os.Args[1])
		log.Error("Select one mode from  'analysmodesis', 'fuzzing' or 'record'")
		err = fmt.Errorf("Unknown mode %s", os.Args[1])
		helper.PrintHelp()
	}

	if err != nil {
		return err
	}

	numberBugs, _, numberTestWithRes, numberErr, numberTimeout := log.GetLoggingNumbers()
	if numberErr == 0 {
		log.Info("Finished with 0 errors")
	} else {
		log.Errorf("Finished with %d errors", numberErr)
	}
	if numberTimeout == 0 {
		log.Info("No internal replay timeouts occurred")
	} else {
		log.Errorf("%d internal replay timeouts occurred", numberTimeout)
	}
	if flags.Mode == "analysis" || flags.Mode == "fuzzing" {
		if numberTestWithRes == 0 {
			log.Info("No bugs have been found/indicated")
		} else {
			if !flags.ModeMain {
				log.Resultf(false, false, "", "Tests with indicated bugs: %d", numberTestWithRes)
			}
			log.Resultf(false, false, "", "Number indicated bugs:  %d", numberBugs)
		}
	}
	timer.UpdateTimeFileOverview("*Total*")

	return nil
}
