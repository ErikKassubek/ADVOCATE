// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief:data for blocking detection
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static"
	"advocate/static/static/s_ssa"
	"advocate/trace"
)

var data *static.Data
var funcCallToSSAFunc map[*trace.ElementFunc]*s_ssa.Function

var blocked map[*trace.ElementAlloc]*trace.Resource

func getSSAFuncFromName(name string) *s_ssa.Function {
	for _, f := range data.Ssa().Funcs() {
		if name == f.Name() {
			return f
		}
	}

	return nil
}
