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
	"advocate/advoc/toolchain"
	"advocate/static/static"
	"advocate/utils/flags"
	"advocate/utils/log"
)

var data *static.Data

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

	IsBlockingBug()

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

func IsBlockingBug() {
	blocked := getBlockedResources()

	for _, alloc := range blocked {
		for _, a := range alloc {
			f, s := data.Ssa().TraceToSSA(a)
			log.Debug(a.ObjID(), " # ", f.Name(), " # ", s.String())
		}
	}

}
