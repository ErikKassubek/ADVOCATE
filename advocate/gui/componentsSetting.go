// Copyright (c) 2026 Erik Kassubek
//
// File: componentsSetting.go
// Brief: Settings
//
// Author: Erik Kassubek
// Created: 2026-06-01
//
// License: BSD-3-Clause

package gui

import (
	"advocate/fuzzing/baseF"
	"advocate/utils/flags"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type componentSetting struct {
	*fyne.Container

	components *settingComponents

	record *settingsRecord
}

func createSettings() componentSetting {
	comp := createSettingComponents()

	cs := componentSetting{
		Container: container.NewVBox(),

		record:     createSettingsRecord(comp),
		components: comp,
	}

	return cs
}

func (self *window) setSettings(content []fyne.CanvasObject) {
	self.settings.Container.Objects = content
	self.settings.Refresh()
}

// ------------------------------------------------------------------------------------
// Setting components
// ------------------------------------------------------------------------------------

type settingComponents struct {
	label          componentSectionLabel
	mainTestSelect componentMainTestSelect

	toRecord  textInput
	toReplay  textInput
	toFuzzing textInput

	maxFuzzingRun     textInput
	maxNumberElements textInput

	measureTime         checkInput
	createStatistics    checkInput
	checkForNotExecuted checkInput

	ignoreCriticalSections checkInput
	ignoreAtomics          checkInput
	onlyAPanicAndLeak      checkInput

	noRewrite   checkInput
	deleteTrace checkInput
	cont        checkInput

	noWarning  checkInput
	noInfo     checkInput
	noProgress checkInput
	output     checkInput

	alwaysPanic checkInput
	noMemorySup checkInput

	fuzzingMode selectInput

	cancelTestIfBugFound checkInput

	scen textInput

	tracePath componentPathSelector
}

func createSettingComponents() *settingComponents {

	to := settingComponents{
		label:                  createSectionLabel("Settings"),
		mainTestSelect:         creatMainTestSelector(),
		toRecord:               createInputNumeric("Timeout Record [s]: ", &flags.TimeoutRecording, true),
		toReplay:               createInputNumeric("Timeout Replay [s]: ", &flags.TimeoutReplay, true),
		toFuzzing:              createInputNumeric("Timeout Fuzzing [s]: ", &flags.TimeoutFuzzing, true),
		maxFuzzingRun:          createInputNumeric("Max. Fuzzing Runs: ", &flags.MaxFuzzingRun, true),
		maxNumberElements:      createInputNumeric("Max. Number Elements: ", &flags.MaxNumberElements, true),
		measureTime:            createInputCheck("Measure Time: ", &flags.MeasureTime),
		createStatistics:       createInputCheck("Create Statistics: ", &flags.CreateStatistics),
		checkForNotExecuted:    createInputCheck("Check for not Executed Ops: ", &flags.NotExecuted),
		ignoreCriticalSections: createInputCheck("Ignore Critical Sections: ", &flags.IgnoreCriticalSection),
		ignoreAtomics:          createInputCheck("Ignore Atomics: ", &flags.IgnoreAtomics),
		onlyAPanicAndLeak:      createInputCheck("Disable Prediction: ", &flags.OnlyAPanicAndLeak),
		noRewrite:              createInputCheck("Skip Rewrite: ", &flags.NoRewrite),
		deleteTrace:            createInputCheck("Delete Trace:", &flags.DeleteTrace),
		cont:                   createInputCheck("Continue: ", &flags.Continue),
		noWarning:              createInputCheck("Disable Warning Massages: ", &flags.NoWarning),
		noInfo:                 createInputCheck("Disable Info Massages: ", &flags.NoWarning),
		noProgress:             createInputCheck("Disable Progress Massages: ", &flags.NoProgress),
		output:                 createInputCheck("Show Program/Test Output: ", &flags.Output),
		alwaysPanic:            createInputCheck("Always Panic: ", &flags.AlwaysPanic),
		noMemorySup:            createInputCheck("Disable Memory Supervisor", &flags.NoMemorySupervisor),
		fuzzingMode:            createInputSelect("Fuzzing Mode: ", &flags.FuzzingMode, baseF.Modes, false),
		cancelTestIfBugFound:   createInputCheck("Cancel Fuzzing If Bug Found: ", &flags.CancelTestIfBugFound),
		scen: createInputTextFunc("Scenations (disable = all): ", "", func(e bool, s string) {
			if e {
				flags.Scenarios = s
			} else {
				flags.Scenarios = ""
			}
			flags.ParseAnalysisCases()
		}, true, true),
		tracePath: createPathSelector("Trace", &flags.TracePath),
	}

	return &to
}
