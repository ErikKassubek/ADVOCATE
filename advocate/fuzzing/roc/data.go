// Copyright (c) 2026
//
// File: data.go
// Brief: Data for guided fuzzing
//
// Created: 2025-10-20
//
// License: BSD-3-Clause

package roc

const (
	maxNumberConstraints    = 10
	maxNumberOfMutsPerConst = 5
	maxTries                = 1000
	lengthConstraint        = 8
	propToSkipEquiv         = 0.9
)

var (
	numberMuts = 0
	traceID    = 0
)
