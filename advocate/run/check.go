// Copyright (c) 2024 Erik Kassubek
//
// File: check.go
// Brief: run some checks
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package run

import (
	"advocate/utils/log"
	"advocate/utils/paths"
	"fmt"
	"os"
	"path/filepath"
)

func Check() error {
	log.Debug("PATH:", paths.GoPatch)

	path := filepath.Join(paths.GoPatch, "bin", "go")
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		pathMake := filepath.Join(paths.GoPatch, "src")
		return fmt.Errorf("Could not find %s. Run make.bash or make.bat in %s before running advocate.", path, pathMake)
	}

	return nil
}
