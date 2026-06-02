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

func createSettings() *componentSetting {
	comp := createSettingComponents()

	cs := &componentSetting{
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

func (self *componentSetting) disable() {
	self.components.disable()
}

func (self *componentSetting) enable() {
	self.components.enable()
}

// ------------------------------------------------------------------------------------
// Setting components
// ------------------------------------------------------------------------------------

type settingComponents struct {
	label          *componentSectionLabel
	mainTestSelect *componentMainTestSelect

	toRecord  *textInput
	toReplay  *textInput
	toFuzzing *textInput

	maxFuzzingRun     *textInput
	maxNumberElements *textInput

	measureTime         *checkInput
	createStatistics    *checkInput
	checkForNotExecuted *checkInput

	ignoreCriticalSections *checkInput
	ignoreAtomics          *checkInput
	onlyAPanicAndLeak      *checkInput

	noRewrite   *checkInput
	deleteTrace *checkInput
	cont        *checkInput

	noWarning  *checkInput
	verbose    *checkInput
	noProgress *checkInput
	output     *checkInput

	alwaysPanic *checkInput
	noMemorySup *checkInput

	fuzzingMode *selectInput

	cancelTestIfBugFound *checkInput

	scen *textInput

	tracePath *componentPathSelector
}

const (
	inverted         = true
	direct           = false
	canBeDisabled    = true
	canNotBeDisabled = false
)

func createSettingComponents() *settingComponents {
	to := settingComponents{
		label:                  createSectionLabel("Settings"),
		mainTestSelect:         creatMainTestSelector(),
		toRecord:               createInputNumeric("Timeout Record [s]: ", &flags.TimeoutRecording, canBeDisabled),
		toReplay:               createInputNumeric("Timeout Replay [s]: ", &flags.TimeoutReplay, canBeDisabled),
		toFuzzing:              createInputNumeric("Timeout Fuzzing [s]: ", &flags.TimeoutFuzzing, canBeDisabled),
		maxFuzzingRun:          createInputNumeric("Max. Fuzzing Runs: ", &flags.MaxFuzzingRun, canBeDisabled),
		maxNumberElements:      createInputNumeric("Max. Number Elements: ", &flags.MaxNumberElements, canBeDisabled),
		measureTime:            createInputCheck("Measure Time: ", &flags.MeasureTime, direct),
		createStatistics:       createInputCheck("Create Statistics: ", &flags.CreateStatistics, direct),
		checkForNotExecuted:    createInputCheck("Check for not Executed Ops: ", &flags.NotExecuted, direct),
		ignoreCriticalSections: createInputCheck("Ignore Critical Sections: ", &flags.IgnoreCriticalSection, direct),
		ignoreAtomics:          createInputCheck("Ignore Atomics: ", &flags.IgnoreAtomics, direct),
		onlyAPanicAndLeak:      createInputCheck("Disable Prediction: ", &flags.OnlyAPanicAndLeak, direct),
		noRewrite:              createInputCheck("Skip Rewrite: ", &flags.NoRewrite, direct),
		deleteTrace:            createInputCheck("Delete Trace:", &flags.DeleteTrace, direct),
		cont:                   createInputCheck("Continue: ", &flags.Continue, direct),
		noWarning:              createInputCheck("Disable Warning Massages: ", &flags.NoWarning, direct),
		verbose:                createInputCheck("Verbose Messages: ", &flags.Verbose, direct),
		noProgress:             createInputCheck("Disable Progress Massages: ", &flags.NoProgress, direct),
		output:                 createInputCheck("Show Program/Test Output: ", &flags.Output, direct),
		alwaysPanic:            createInputCheck("Always Panic: ", &flags.AlwaysPanic, direct),
		noMemorySup:            createInputCheck("Disable Memory Supervisor", &flags.NoMemorySupervisor, direct),
		fuzzingMode:            createInputSelect("Fuzzing Mode: ", &flags.FuzzingMode, baseF.Modes, canNotBeDisabled),
		cancelTestIfBugFound:   createInputCheck("Cancel Fuzzing If Bug Found: ", &flags.CancelTestIfBugFound, direct),
		scen: createInputTextFunc("Scenations (disabled = all): ", "",
			func(e bool, s string) {
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

func (self *settingComponents) disable() {
	self.mainTestSelect.disable()

	self.toRecord.disable()
	self.toReplay.disable()
	self.toFuzzing.disable()

	self.maxFuzzingRun.disable()
	self.maxNumberElements.disable()

	self.measureTime.disable()
	self.createStatistics.disable()
	self.checkForNotExecuted.disable()

	self.ignoreCriticalSections.disable()
	self.ignoreAtomics.disable()
	self.onlyAPanicAndLeak.disable()

	self.noRewrite.disable()
	self.deleteTrace.disable()
	self.cont.disable()

	self.noWarning.disable()
	self.verbose.disable()
	self.noProgress.disable()
	self.output.disable()

	self.alwaysPanic.disable()
	self.noMemorySup.disable()

	self.fuzzingMode.disable()

	self.cancelTestIfBugFound.disable()

	self.scen.disable()

	self.tracePath.disable()

	win.modeSelect.disable()
	win.projSelector.disable()
}

func (self *settingComponents) enable() {
	self.mainTestSelect.enable()

	self.toRecord.enable()
	self.toReplay.enable()
	self.toFuzzing.enable()

	self.maxFuzzingRun.enable()
	self.maxNumberElements.enable()

	self.measureTime.enable()
	self.createStatistics.enable()
	self.checkForNotExecuted.enable()

	self.ignoreCriticalSections.enable()
	self.ignoreAtomics.enable()
	self.onlyAPanicAndLeak.enable()

	self.noRewrite.enable()
	self.deleteTrace.enable()
	self.cont.enable()

	self.noWarning.enable()
	self.verbose.enable()
	self.noProgress.enable()
	self.output.enable()

	self.alwaysPanic.enable()
	self.noMemorySup.enable()

	self.fuzzingMode.enable()

	self.cancelTestIfBugFound.enable()

	self.scen.enable()

	self.tracePath.enable()
}
