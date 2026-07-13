// Copyright (c) 2026 Erik Kassubek
//
// File: blocking.go
// Brief: Entry point for static blocking analysis
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static"
	"advocate/utils/flags"
)

var data *static.Data

// Main function for testing static analysis.
// Todo: remove when static analysis is fully implemented
func Test() {
	BuildStaticBlockingAnalysis(flags.ProgPath)
}

// init to static blocking analysis
func BuildStaticBlockingAnalysis(dir string) (err error) {

	data, err = static.BuildStaticData(dir)
	if err != nil {
		return err
	}

	return nil
}

func IsBlockingBug() {
	_ = getBlockedResources()

}
