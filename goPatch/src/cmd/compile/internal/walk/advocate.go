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
func shouldAdvocate(n ir.Node) bool {
	if n == nil {
		return false
	}

	if sym := nodeSym(n); sym != nil && sym.Pkg != nil {
		return !isStdPkgPath(sym.Pkg.Path)
	}

	if !n.Pos().IsKnown() {
		return false // synthetic/unknown position — be conservative
	}

	pkgPath := types.LocalPkg.Path // package currently being compiled
	if isStdPkgPath(pkgPath) {
		return false
	}

	return !isAdvocateCall(n)
}

func nodeSym(n ir.Node) *types.Sym {
	switch x := ir.StaticValue(n).(type) {
	case *ir.Name:
		return x.Sym()
	case *ir.SelectorExpr:
		if x.Selection != nil {
			return x.Selection.Sym
		}
		return x.Sel
	case *ir.CallExpr:
		return nodeSym(x.Fun)
	}
	if t := n.Type(); t != nil && t.Sym() != nil {
		return t.Sym()
	}
	return nil
}

func isStdPkgPath(pkgPath string) bool {
	if pkgPath == "" || pkgPath == "main" {
		return true
	}
	first := pkgPath
	if i := strings.IndexByte(pkgPath, '/'); i >= 0 {
		first = pkgPath[:i]
	}
	return !strings.Contains(first, ".")
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
	sh := shouldAdvocate(fn)
	println(base.Ctxt.PosTable.Pos(fn.Pos()).Base().AbsFilename(), " ", sh)

	if !sh {
		return
	}

	fn.Body = instrumentStmtList(fn.Body)
}

func instrumentStmtList(body ir.Nodes) ir.Nodes {
	out := make(ir.Nodes, 0, len(body)*2)

	for _, stmt := range body {
		// First recurse into children
		instrumentStmtRecursive(stmt)

		out.Append(stmt)

		if isAdvocateCall(stmt) {
			continue
		}

		if n := addAlloc(stmt); n != nil {
			out.Append(n)
		}

		if n := addIfSwitch(stmt); n != nil {
			out.Append(n)
		}
	}

	return out
}

func instrumentStmtRecursive(n ir.Node) {
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

func addIfSwitch(n ir.Node) ir.Node {
	switch n := n.(type) {
	case *ir.IfStmt:
		walkIfChain(n)
	}

	return nil
}

func walkIfChain(n *ir.IfStmt) {
	for n != nil {
		// process n.Cond, n.Body here
		fmt.Printf("if %v { ... }\n", n.Cond)

		if len(n.Else) == 1 {
			if elseif, ok := n.Else[0].(*ir.IfStmt); ok {
				// this is an "else if"
				n = elseif
				continue
			}
		}

		// plain else block (0 or >1 statements, or a single non-If statement)
		if len(n.Else) > 0 {
			fmt.Printf("else { %v }\n", n.Else)
		}
		break
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
	// TODO: for now only record channels
	return nil, nil

	// switch n := n.(type) {

	// case *ir.AssignStmt:
	// 	// m := sync.Mutex{}
	// 	if n.Y != nil {
	// 		t = n.Y.Type()
	// 	}
	// 	name, ok := n.X.(*ir.Name)
	// 	if ok {
	// 		m = name
	// 	}

	// case *ir.Decl:
	// 	// var m sync.Mutex
	// 	if n.X != nil {
	// 		t = n.X.Type()
	// 		m = n.X
	// 	}
	// }

	// return
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
