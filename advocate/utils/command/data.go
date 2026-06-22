// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: store data
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package command

import "context"

const (
	NoTimeout = -1
	NoDir     = ""
)

var (
	Ctx context.Context
)
