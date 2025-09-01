// Copyright (c) 2025 Erik Kassubek
//
// File: flags.go
// Brief: Command line flags
//
// Author: Erik Kassubek
// Created: 2025-08-26
//
// License: BSD-3-Clause

package flags

// Paths and names
var (
	ProgPath  string
	TracePath string

	ProgName string
	ExecName string
)

// Modes
var (
	ModeMain    bool
	FuzzingMode string
)

// timeouts and limits
var (
	TimeoutRecording int
	TimeoutReplay    int
	TimeoutFuzzing   int
	MaxFuzzingRun    int

	MaxNumberElements int
)

// logging
var (
	Output     bool
	NoInfo     bool
	NoProgress bool
	NoWarning  bool
)

// statistics
var (
	MeasureTime      bool
	CreateStatistics bool
	NotExecuted      bool
)

// memory and panic
var (
	NoMemorySupervisor bool
	AlwaysPanic        bool
)

// settings
var (
	IgnoreAtomics         bool
	IgnoreCriticalSection bool
	IgnoreFifo            bool

	OnlyAPanicAndLeak bool

	CancelTestIfBugFound bool

	Settings string

	Scenarios string
)

// execution control
var (
	Continue      bool
	SkipExisting  bool
	NoRewrite     bool
	NoSkipRewrite bool
	KeepTraces    bool
)
