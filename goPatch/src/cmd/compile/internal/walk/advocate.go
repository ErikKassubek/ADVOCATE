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
	"cmd/compile/internal/types"
	"cmd/internal/src"
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

	// Never instrument the compiler/runtime/toolchain itself.
	switch {
	case pkg == "runtime":
		return false

	case pkg == "runtime/internal":
		return false

	case pkg == "syscall":
		return false

	case pkg == "os":
		return false

	case pkg == "internal/syscall":
		return false

	case strings.HasPrefix(pkg, "internal/"):
		return false

	case strings.HasPrefix(pkg, "cmd/"):
		return false

	case strings.HasPrefix(pkg, "bootstrap/"):
		return false

	case strings.HasPrefix(pkg, "go/"):
		return false
	}

	// Never instrument compiler source files.
	if fn.Pos().IsKnown() {
		file := base.Ctxt.PosTable.Pos(fn.Pos()).Filename()

		// Your modified Go tree
		if strings.Contains(file, "/goPatch/") {
			return false
		}

		// Any GOROOT source tree
		if strings.Contains(file, "/src/cmd/") ||
			strings.Contains(file, "/src/runtime/") ||
			strings.Contains(file, "/src/internal/") {
			return false
		}
	}

	// Compiler generated wrappers and special functions
	if fn.Pragma&ir.Nosplit != 0 || fn.Wrapper() {
		return false
	}

	name := fn.Sym().Name

	switch name {
	case "advocateFunctionCall",
		"advocateFunctionReturn",
		"AdvocateAllocMutex",
		"AdvocateAllocCondVar",
		"AdvocateAllocWG":
		return false
	}

	return true
}

func isAdvocateCall(n ir.Node) bool {
	call, ok := n.(*ir.CallExpr)
	if !ok {
		return false
	}

	if call.Op() != ir.OCALLFUNC {
		return false
	}

	name, ok := call.Fun.(*ir.Name)
	if !ok {
		return false
	}

	return name.Sym() != nil &&
		(name.Sym().Name == "AdvocateAllocMutex" ||
			name.Sym().Name == "AdvocateAllocCondVar" ||
			name.Sym().Name == "AdvocateAllocWG")
}

func instrumentBody(fn *ir.Func) {
	if !shouldAdvocate(fn) {
		return
	}

	fn.Body = instrumentStmtList(fn.Body)
}

func instrumentStmtList(body ir.Nodes) ir.Nodes {
	out := make(ir.Nodes, 0, len(body)*2)

	for _, stmt := range body {
		// First recurse into children
		instrumentStmt(stmt)

		out.Append(stmt)

		if isAdvocateCall(stmt) {
			continue
		}

		if n := addAlloc(stmt); n != nil {
			out.Append(n)
		}
	}

	return out
}

func instrumentStmt(n ir.Node) {
	switch n := n.(type) {

	case *ir.BlockStmt:
		n.List = instrumentStmtList(n.List)

	case *ir.IfStmt:
		n.Body = instrumentStmtList(n.Body)

		if n.Else != nil {
			n.Else = instrumentStmtList(n.Else)
		}

	case *ir.ForStmt:
		n.Body = instrumentStmtList(n.Body)

	case *ir.RangeStmt:
		n.Body = instrumentStmtList(n.Body)

	case *ir.SwitchStmt:
		for _, c := range n.Compiled {
			switch c := c.(type) {
			case *ir.CaseClause:
				c.Body = instrumentStmtList(c.Body)
			}
		}

	case *ir.SelectStmt:
		for _, c := range n.Compiled {
			switch c := c.(type) {
			case *ir.CommClause:
				c.Body = instrumentStmtList(c.Body)
			}
		}
	}
}

// TODO: variables created in init
func addAlloc(n ir.Node) ir.Node {
	t, obj := allocatedSync(n)
	if t == nil || obj == nil {
		return nil
	}

	var runtimeName string

	switch {
	case isSyncType(t, "Mutex") || isSyncType(t, "RWMutex"):
		fmt.Println(n)
		runtimeName = "AdvocateAllocMutex"

	case isSyncType(t, "Cond"):
		runtimeName = "AdvocateAllocCondVar"

	case isSyncType(t, "WaitGroup"):
		runtimeName = "AdvocateAllocWG"

	default:
		return nil
	}

	fn := typecheck.LookupRuntime(runtimeName)

	addr := ir.NewAddrExpr(
		n.Pos(),
		obj,
	)

	addr.SetType(types.NewPtr(obj.Type()))
	addr.SetTypecheck(1)

	unsafeAddr := makeUnsafePointer(addr, n.Pos())

	call := typecheck.Call(
		n.Pos(),
		fn,
		[]ir.Node{unsafeAddr},
		false,
	)

	return call
}

func makeUnsafePointer(addr ir.Node, pos src.XPos) ir.Node {
	ptrType := types.Types[types.TUNSAFEPTR]

	return ir.NewConvExpr(
		pos,
		ir.OCONV,
		ptrType,
		addr,
	)
}

func allocatedSync(n ir.Node) (t *types.Type, m *ir.Name) {
	switch n := n.(type) {

	case *ir.AssignStmt:
		// m := sync.Mutex{}
		if n.Y != nil {
			t = n.Y.Type()
		}
		name, ok := n.X.(*ir.Name)
		if ok {
			m = name
		}

	case *ir.Decl:
		// var m sync.Mutex
		if n.X != nil {
			t = n.X.Type()
			m = n.X
		}
	}

	return
}

func isSyncType(t *types.Type, name string) bool {
	if t == nil {
		return false
	}

	sym := t.Sym()
	if sym == nil {
		return false
	}

	return sym.Pkg != nil &&
		sym.Pkg.Path == "sync" &&
		sym.Name == name
}
