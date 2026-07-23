// Copyright (c) 2026 Erik Kassubek
//
// File: blocking.go
// Brief: Entry point for static blocking analysis
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/advoc/toolchain"
	"advocate/analysis/a_base"
	"advocate/static/static"
	"advocate/utils/flags"
	"advocate/utils/log"
)

// init to static blocking analysis
func BuildStaticBlockingAnalysis() (err error) {
	log.Info("Build static Analysis")

	il := 0
	if flags.ModeMain {
		il, err = toolchain.HeaderInsertDummyMain()
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		log.Error("Static not implemented for tests")
		// TODO: implement
	}

	data, err = static.BuildStaticData(flags.RootPath)

	if flags.ModeMain {
		toolchain.HeaderRemoverDummyMain(il)
	} else {
		log.Error("Static not implemented for tests")
		// TODO: implement
	}

	if err != nil {
		return err
	}

	isBlockingBug()

	data.Ssa().PrintAnalysis()
	// data.Ssa().PrintSsa(true)

	// tr := a_base.MainTrace.AsIterator()
	// for elem := tr.Next(); elem != nil; elem = tr.Next() {
	// 	f, instr := data.Ssa().TraceToSSA(elem)
	// 	if instr != nil {
	// 		println(f.Name(), " # ", elem.ToString(), " # ", instr.String())
	// 	}
	// }

	return nil
}

func isBlockingBug() {
	blocked := getBlockedResources()

	for r := range blocked {
		f, s := data.Ssa().TraceToSSA(r.Alloc())
		log.Debug(r.Id(), " # ", f.Name(), " # ", s.String())
	}

	buildFuncCallToSSAFunc()
	log.Debug(a_base.MainTrace.CallTree().String())

	// determineResouceToSSAAtTermination()
}
