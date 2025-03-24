// Copyright (c) 2024 Erik Kassubek
//
// File: headerUnitTests.go
// Brief: Functions to add and remove the ADVOCATE header into file containing
//    unit tests
//
// Author: Erik Kassubek
// Created: 2024-10-29
//
// License: BSD-3-Clause

package toolchain

import (
	"io"
	"os"
	"strings"
)

// extractTraceNumber extracts the numeric part from a trace directory name
func extractTraceNumber(trace string) string {
	parts := strings.Split(trace, "rewrittenTrace_")
	if len(parts) > 1 {
		return parts[1]
	}
	parts = strings.Split(trace, "advocateTraceReplay_")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func checkForTimeout(output string) bool {
	outFile, err := os.Open(output)
	if err != nil {
		return false
	}
	defer outFile.Close()

	content, err := io.ReadAll(outFile)
	if err != nil {
		return false
	}

	if strings.Contains(string(content), "panic: test timed out after") {
		return true
	}

	return false
}
