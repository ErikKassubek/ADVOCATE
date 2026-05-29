package app

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
func Run() {

	// If -main is set, the path needs to be the path to the main file
	// If the given path is to a folder, check if a main.go file exists in this folder
	// If so, fix the path. Otherwise return error and finish
	if flags.ModeMain {
		var err error
		flags.ProgPath, err = paths.GetMainPath(flags.ProgPath)
		if err != nil {
			log.Error("Could not find main file. If -main is set, -path should point to the main file.")
			log.Error(err)

			return
		}
	}

	settings.SetSettings()
	paths.BuildPaths(flags.ModeMain)

	progPathDir := paths.GetDirectory(flags.ProgPath)
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

	var err error
	baseA.AnalysisCasesMap, err = flags.ParseAnalysisCases()
	if err != nil {
		log.Error("Could not read analysis cases: ", err)
		return
	}

	modeMainTest := "test"
	if flags.ModeMain {
		modeMainTest = "main"
	}

	helper.CheckGoMod()
	helper.RunGoModTidy()

	if flags.ModeMain && flags.ExecName == "" {
		log.Error("Could not determine executable name from go.mod. Provide with -exec [ExecutableName]")
		panic(fmt.Errorf("Could not determine executable name"))
	}

	switch flags.Mode {
	case "analysis":
		modeToolchain(modeMainTest, true, true, true)
	case "fuzzing":
		modeFuzzing()
	case "record", "recording":
		flags.KeepTraces = true
		modeToolchain(modeMainTest, true, false, false)
	case "replay":
		modeToolchain(modeMainTest, false, false, true)
	default:
		log.Errorf("Unknown mode %s\n", os.Args[1])
		log.Error("Select one mode from  'analysis', 'fuzzing' or 'record'")
		helper.PrintHelp()
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
}
