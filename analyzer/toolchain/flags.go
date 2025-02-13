// Copyright (c) 2025 Erik Kassubek
//
// File: flags.go
// Brief: Store the flags needed in runAnalyzer
//
// Author: Erik Kassubek
// Created: 2025-02-04
//
// License: BSD-3-Clause

package toolchain

var (
	noRewriteFlag             bool
	analyisCasesFlag          map[string]bool
	ignoreAtomicsFlag         bool
	fifoFlag                  bool
	ignoreCriticalSectionFlag bool
	rewriteAllFlag            bool
	ignoreRewriteFlag         string
	onlyAPanicAndLeakFlag     bool
)

func SetFlags(noRewrite bool, analyisCases map[string]bool, ignoreAtomics,
	fifo, ignoreCriticalSection, rewriteAll bool, ignoreRewrite string, onlyAPanicAndLeak bool) {

	noRewriteFlag = noRewrite

	analyisCasesFlag = analyisCases

	ignoreAtomicsFlag = ignoreAtomics
	fifoFlag = fifo
	ignoreCriticalSectionFlag = ignoreCriticalSection
	rewriteAllFlag = rewriteAll
	ignoreRewriteFlag = ignoreRewrite
	onlyAPanicAndLeakFlag = onlyAPanicAndLeak
}
