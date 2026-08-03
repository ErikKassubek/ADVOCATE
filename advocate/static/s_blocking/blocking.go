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
	"advocate/static/static"
	"advocate/utils/flags"
	"advocate/utils/log"
)

var data *static.Data

// init to static blocking analysis
func BuildStaticBlockingAnalysis() (err error) {
	log.Info("Build static Analysis")

	data, err = static.BuildStaticData(flags.RootPath)
	if err != nil {
		return err
	}

	data.Ssa().PrintAnalysis()
	// data.Ssa().PrintSsa(true)

	isBlockingBug()

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
	data.Blocking().Blocked = getBlockedResources()

	for _, r := range data.Blocking().Blocked {
		f, s := data.Ssa().TraceToSSA(r.Alloc())

		if f == nil || s == nil {
			continue
		}

		log.Debug(r.Id(), " # ", f.Name(), " # ", s.String())
	}
	log.Debug("")

	buildFuncCallToSSAFunc()

	determineResouceToSSAAtTermination()
}
