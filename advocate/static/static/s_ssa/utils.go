// Copyright (c) 2026 Erik Kassubek
//
// File: utils.go
// Brief: Helper for ssa
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

type InstClass string

const (
	Ic_unknown             InstClass = "unknown"
	Ic_alloc               InstClass = "alloc"
	Ic_binOp               InstClass = "binOp"
	Ic_call                InstClass = "call"
	Ic_changeInterface     InstClass = "changeInterface"
	Ic_changeType          InstClass = "changeType"
	Ic_const               InstClass = "const"
	Ic_convert             InstClass = "convert"
	Ic_debugRef            InstClass = "debugRef"
	Ic_defer               InstClass = "defer"
	Ic_extract             InstClass = "extract"
	Ic_field               InstClass = "field"
	Ic_fieldAddr           InstClass = "fieldAddr"
	Ic_freeVar             InstClass = "freeVar"
	Ic_function            InstClass = "function"
	Ic_go                  InstClass = "go"
	Ic_if                  InstClass = "if"
	Ic_index               InstClass = "index"
	Ic_indexAddr           InstClass = "indexAddr"
	Ic_jump                InstClass = "jump"
	Ic_lookup              InstClass = "lookup"
	Ic_makeChan            InstClass = "makeChan"
	Ic_makeClosure         InstClass = "makeClosure"
	Ic_makeInterface       InstClass = "makeInterface"
	Ic_makeMap             InstClass = "makeMap"
	Ic_makeSlice           InstClass = "makeSlice"
	Ic_mapUpdate           InstClass = "mapUpdate"
	Ic_multiConvert        InstClass = "multiConvert"
	Ic_next                InstClass = "next"
	Ic_panic               InstClass = "panic"
	Ic_parameter           InstClass = "parameter"
	Ic_phi                 InstClass = "phi"
	Ic_range               InstClass = "range"
	Ic_return              InstClass = "return"
	Ic_runDefers           InstClass = "runDefers"
	Ic_select              InstClass = "select"
	Ic_send                InstClass = "send"
	Ic_slice               InstClass = "slice"
	Ic_sliceToArrayPointer InstClass = "sliceToArrayPointer"
	Ic_store               InstClass = "store"
	Ic_typeAssert          InstClass = "typeAssert"
	Ic_unOp                InstClass = "unOp"
)

type hasConcInfo [4]bool

type concRes int

const (
	chanInd concRes = iota
	mutexInd
	condVarInd
	wgInd
)

func (this hasConcInfo) Resource() bool {
	for i := 0; i < 4; i++ {
		if this[i] {
			return true
		}
	}

	return false
}

func (this hasConcInfo) Channel() bool {
	return this[chanInd]
}

func (this hasConcInfo) Mutex() bool {
	return this[mutexInd]
}

func (this hasConcInfo) CondVar() bool {
	return this[condVarInd]
}

func (this hasConcInfo) WaitGroup() bool {
	return this[wgInd]
}

func (this *concRes) string() string {
	switch *this {
	case chanInd:
		return "chan"
	case mutexInd:
		return "mutex"
	case condVarInd:
		return "condVar"
	case wgInd:
		return "wg"
	}
	return ""
}
