// ADVOCATE-FILE-START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate.go
// Brief: Insert recording for mutex, cond var and wait group creation
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package walk

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"fmt"
	"strings"
)

// ADOVCATE-START
// MARK: shouldAdvocate
func shouldAdvocate(fn *ir.Func) bool {
	if fn == nil || fn.Sym() == nil || fn.Sym().Pkg == nil {
		return false
	}

	pkg := fn.Sym().Pkg.Path

	if fn.Pos().IsKnown() {
		file := base.Ctxt.PosTable.Pos(fn.Pos()).Filename()
		if strings.Contains(file, "/goPatch/") {
			return false
		}
	}

	if pkg == "runtime" ||
		pkg == "syscall" ||
		pkg == "os" ||
		pkg == "internal/syscall" {

		return false
	}

	if fn.Pragma&ir.Nosplit != 0 || fn.Wrapper() {
		return false
	}

	name := fn.Sym().Name

	if name == "advocateFunctionCall" || name == "advocateFunctionReturn" {
		return false
	}

	return true
}

func instrumentBody(fn *ir.Func) {
	if !shouldAdvocate(fn) {
		return
	}

	old := fn.Body
	out := make(ir.Nodes, 0, len(old))

	for _, stmt := range old {
		out.Append(stmt)

		if n := addAlloc(stmt); n != nil {
			out.Append(n)
		}
	}

	fn.Body = out
}

func addAlloc(n ir.Node) ir.Node {
	if !isMutexCreation(n) {
		return nil
	}

	switch n.(type) {
	case *ir.AssignStmt:
		fmt.Println(n)

		fn := typecheck.LookupRuntime("AdvocateAllocMutex")

		call := ir.NewCallExpr(
			n.Pos(),
			ir.OCALLFUNC,
			fn,
			nil,
		)

		call.SetTypecheck(1)

		return call
	}

	return nil

}

func isMutexCreation(n ir.Node) bool {
	as, ok := n.(*ir.AssignStmt)
	if !ok {
		return false
	}

	// Only handle single assignment: m := sync.Mutex{}
	rhs := as.Y
	if rhs == nil {
		return false
	}

	t := rhs.Type()
	if t == nil {
		return false
	}

	// Strip aliases/named wrappers as needed.
	sym := t.Sym()
	if sym == nil {
		return false
	}

	return sym.Pkg != nil &&
		sym.Pkg.Path == "sync" &&
		sym.Name == "Mutex"
}

// ADVOCATE-FILE-END
