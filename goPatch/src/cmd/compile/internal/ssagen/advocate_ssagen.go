// ADVOCATE-FILE-START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_ssagen.go
// Brief: Insert recording for mutex, cond var and wait group creation
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package ssagen

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
)

func syncObjectType(t *types.Type) string {
	if t == nil {
		return ""
	}

	sym := t.Sym()
	if sym == nil || sym.Pkg == nil {
		return ""
	}

	if sym.Pkg.Path != "sync" {
		return ""
	}

	switch sym.Name {
	case "Mutex":
		return "M"

	case "Cond":
		return "D"

	case "WaitGroup":
		return "W"
	}

	return ""
}

func syncObjectDecl(n ir.Node) string {
	switch n := n.(type) {

	case *ir.Decl:
		// var m sync.Mutex
		// var c sync.Cond
		// var w sync.WaitGroup

		return syncObjectType(n.X.Type())

	case *ir.AssignStmt:
		// m := sync.Mutex{}
		// c := sync.Cond{}
		// w := sync.WaitGroup{}

		if lhs, ok := n.X.(*ir.Name); ok {
			return syncObjectType(lhs.Type())
		}
	}

	return ""
}

func addAllocToBody(fn *ir.Func) {
	newBody := make(ir.Nodes, 0, len(fn.Body))

	for _, n := range fn.Body {
		newBody.Append(n)

		obj := syncObjectDecl(n)

		switch obj {
		case "M":
			newBody.Append(makeAdvocateAllocCall("AdvocateAllocMutex"))
		case "D":
			newBody.Append(makeAdvocateAllocCall("AdvocateAllocCondVar"))
		case "W":
			newBody.Append(makeAdvocateAllocCall("AdvocateAllocWGr"))
		}
	}

	fn.Body = newBody
}

func makeAdvocateAllocCall(name string) ir.Node {
	fn := typecheck.LookupRuntime(name)

	return ir.NewCallExpr(
		base.Pos,
		ir.OCALLFUNC,
		fn,
		nil,
	)
}

// ADVOCATE-FILE-END
