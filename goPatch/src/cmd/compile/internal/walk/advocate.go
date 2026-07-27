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
	"strings"
)

// ADOVCATE-START
// MARK: shouldAdvocate
func shouldAdvocate(n ir.Node) bool {
	if n == nil {
		return false
	}

	if !(base.Flag.AdvocateTrace || base.Flag.AdvocateReplay || base.Flag.AdvocateFuzzing) {
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
	if !shouldAdvocate(fn) {
		return
	}

	fn.Body = instrumentStmtList(fn.Body)
}

func instrumentStmtList(body ir.Nodes) ir.Nodes {
	out := make(ir.Nodes, 0, len(body)*2)

	for _, stmt := range body {
		instrumentStmtRecursive(stmt)

		out.Append(stmt)

		if n := addAlloc(stmt); n != nil {
			out.Append(n)
		}
	}

	return out
}

func instrumentStmtRecursive(n ir.Node) {
	if n == nil {
		return
	}

	switch x := n.(type) {
	case *ir.BlockStmt:
		x.List = instrumentStmtList(x.List)

	case *ir.IfStmt:
		x.Cond = instrumentExprRecursive(x.Cond)
		x.Body = instrumentStmtList(x.Body)
		if x.Else != nil {
			x.Else = instrumentStmtList(x.Else)
		}

	case *ir.ForStmt:
		x.Cond = instrumentExprRecursive(x.Cond)
		x.Body = instrumentStmtList(x.Body)

	case *ir.RangeStmt:
		x.X = instrumentExprRecursive(x.X)
		x.Body = instrumentStmtList(x.Body)

	case *ir.SwitchStmt:
		for _, c := range x.Compiled {
			if c, ok := c.(*ir.CaseClause); ok {
				c.Body = instrumentStmtList(c.Body)
			}
		}

	case *ir.SelectStmt:
		for _, c := range x.Compiled {
			if c, ok := c.(*ir.CommClause); ok {
				c.Body = instrumentStmtList(c.Body)
			}
		}

	default:
		// Let expression traversal handle assignments, calls, literals, etc.
		instrumentExprRecursive(n)
	}
}

func instrumentExprRecursive(n ir.Node) ir.Node {
	if n == nil {
		return nil
	}

	switch n := n.(type) {

	case *ir.CallExpr:
		n.Fun = instrumentExprRecursive(n.Fun)

		for i, arg := range n.Args {
			n.Args[i] = instrumentExprRecursive(arg)
		}

	case *ir.BinaryExpr:
		n.X = instrumentExprRecursive(n.X)
		n.Y = instrumentExprRecursive(n.Y)

	case *ir.UnaryExpr:
		n.X = instrumentExprRecursive(n.X)

	case *ir.ConvExpr:
		n.X = instrumentExprRecursive(n.X)

	case *ir.AddrExpr:
		n.X = instrumentExprRecursive(n.X)

	case *ir.SelectorExpr:
		n.X = instrumentExprRecursive(n.X)

	case *ir.IndexExpr:
		n.X = instrumentExprRecursive(n.X)
		n.Index = instrumentExprRecursive(n.Index)

	case *ir.SliceExpr:
		n.X = instrumentExprRecursive(n.X)
		n.Low = instrumentExprRecursive(n.Low)
		n.High = instrumentExprRecursive(n.High)
		n.Max = instrumentExprRecursive(n.Max)

	case *ir.CompLitExpr:
		for i, e := range n.List {
			n.List[i] = instrumentExprRecursive(e)
		}

	case *ir.StructKeyExpr:
		n.Value = instrumentExprRecursive(n.Value)

	}

	return n
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

		if len(n.Else) == 1 {
			if elseif, ok := n.Else[0].(*ir.IfStmt); ok {
				// this is an "else if"
				n = elseif
				continue
			}
		}

		break
	}
}

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

	return typecheck.Call(
		n.Pos(),
		fn,
		[]ir.Node{unsafeAddr},
		false,
	)
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

func allocatedSync(n ir.Node) (*types.Type, ir.Node) {
	switch x := n.(type) {

	case *ir.AssignStmt:
		// m := sync.Mutex{}
		if x.X == nil {
			return nil, nil
		}

		return x.X.Type(), x.X

	case *ir.Decl:
		// var m sync.Mutex
		if x.X == nil {
			return nil, nil
		}

		return x.X.Type(), x.X
	}

	return nil, nil
}

// func isSyncType(t *types.Type, name string) bool {
// 	if t == nil {
// 		return false
// 	}

// 	sym := t.Sym()
// 	if sym == nil {
// 		return false
// 	}

// 	if sym.Pkg.Path == "sync" {
// 		fmt.Println(sym.Pkg.Path, sym.Name)
// 	}

// 	return sym.Pkg != nil &&
// 		sym.Pkg.Path == "sync" &&
// 		sym.Name == name
// }

func isSyncType(t *types.Type, name string) bool {
	if t == nil {
		return false
	}

	sym := t.Sym()
	if sym == nil || sym.Pkg == nil {
		return false
	}

	return sym.Pkg.Path == "sync" &&
		sym.Name == name
}
