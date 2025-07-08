// Copyright (c) 2025 Erik Kassubek
//
// File: hb.go
// Brief: Happens before enum
//
// Author: Erik Kassubek
// Created: 2025-07-08
//
// License: BSD-3-Clause

package hb

type HappensBefore int

const (
	Before = iota
	After
	Concurrent
	None
)
