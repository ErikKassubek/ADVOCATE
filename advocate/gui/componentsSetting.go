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
}

func createSettingComponents() *settingComponents {
	to := settingComponents{
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
	}

	return &to
}
