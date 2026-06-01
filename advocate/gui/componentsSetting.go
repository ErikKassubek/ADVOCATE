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
	toRecord          textInput
	toReplay          textInput
	toFuzzing         textInput
	maxFuzzingRun     textInput
	maxNumberElements textInput
}

func createSettingComponents() *settingComponents {
	to := settingComponents{
		toRecord:          createNumericInput("Timeout Record [s]: ", &flags.TimeoutRecording, true),
		toReplay:          createNumericInput("Timeout Replay [s]: ", &flags.TimeoutReplay, true),
		toFuzzing:         createNumericInput("Timeout Fuzzing [s]: ", &flags.TimeoutFuzzing, true),
		maxFuzzingRun:     createNumericInput("Max. Fuzzing Runs: ", &flags.MaxFuzzingRun, true),
		maxNumberElements: createNumericInput("Max. Number Elements: ", &flags.MaxNumberElements, true),
	}

	return &to
}
