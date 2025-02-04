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
	noPrintFlag               bool
	noRewriteFlag             bool
	analyisCasesFlag          map[string]bool
	ignoreAtomicsFlag         bool
	fifoFlag                  bool
	ignoreCriticalSectionFlag bool
	noWarningFlag             bool
	rewriteAllFlag            bool
	ignoreRewriteFlag         string
)

func SetFlags(noPrint, noRewrite, noWarning bool, analyisCases map[string]bool, ignoreAtomics,
	fifo, ignoreCriticalSection, rewriteAll bool, ignoreRewrite string) {

	noPrintFlag = noPrint
	noRewriteFlag = noRewrite
	noWarningFlag = noWarning

	analyisCasesFlag = analyisCases

	ignoreAtomicsFlag = ignoreAtomics
	fifoFlag = fifo
	ignoreCriticalSectionFlag = ignoreCriticalSection
	rewriteAllFlag = rewriteAll
	ignoreRewriteFlag = ignoreRewrite
}
