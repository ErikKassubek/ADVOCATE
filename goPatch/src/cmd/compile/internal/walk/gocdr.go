// GOCDR-FILE-START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr.go
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

// ==================================================
// MARK: instrument
// ==================================================

func instrumentBody(fn *ir.Func) {
	isGoCDR := base.Flag.GoCDRTrace || base.Flag.GoCDRReplay || base.Flag.GoCDRFuzzing
	if !isGoCDR {
		return
	}

	if isUserMain(fn) {
		fn.Body = addInit(fn.Body, fn.Pos())
	}

	if !shouldGoCDR(fn) {
		return
	}

	instrumentParameterCopy(fn)

	fn.Body = instrumentStmtList(fn.Body)
}

func addInit(body ir.Nodes, pos src.XPos) ir.Nodes {
	out := make(ir.Nodes, 0)

	if base.Flag.GoCDRTrace {
		fn := typecheck.LookupRuntime("GoCDRInitTracing")
		out.Append(typecheck.Call(
			pos,
			fn,
			[]ir.Node{
				ir.NewInt(pos, int64(base.Flag.GoCDRTimeout)),
				ir.NewBool(pos, false),
			},
			false,
		))
	} else if base.Flag.GoCDRReplay {
		fn := typecheck.LookupRuntime("GoCDRInitReplay")
		out.Append(typecheck.Call(
			pos,
			fn,
			[]ir.Node{
				ir.NewString(pos, base.Flag.GoCDRPath),
				ir.NewInt(pos, int64(base.Flag.GoCDRTimeout)),
				ir.NewBool(pos, base.Flag.GoCDRAtomics),
				ir.NewBool(pos, true),
			},
			false,
		))
	} else if base.Flag.GoCDRFuzzing {
		fn := typecheck.LookupRuntime("GoCDRInitFuzzing")
		out.Append(typecheck.Call(
			pos,
			fn,
			[]ir.Node{
				ir.NewString(pos, base.Flag.GoCDRPath),
				ir.NewInt(pos, int64(base.Flag.GoCDRTimeout)),
				ir.NewBool(pos, false),
			},
			false,
		))
	} else {
		return body
	}

	out.Append(body...)

	return out
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
		instrumentIfChain(x)

	case *ir.ForStmt:
		x.Cond = instrumentExprRecursive(x.Cond)
		x.Body = instrumentStmtList(x.Body)

	case *ir.RangeStmt:
		x.X = instrumentExprRecursive(x.X)
		x.Body = instrumentStmtList(x.Body)

	case *ir.SwitchStmt:
		instrumentSwitch(x)

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

// ==================================================
// MARK: Alloc
// ==================================================

func addAlloc(n ir.Node) ir.Node {
	t, obj := allocatedSync(n)
	if t == nil || obj == nil {
		return nil
	}

	var runtimeName string

	switch {
	case isSyncType(t, "Mutex") || isSyncType(t, "RWMutex"):
		runtimeName = "GoCDRAllocMutex"

	case isSyncType(t, "Cond"):
		runtimeName = "GoCDRAllocCondVar"

	case isSyncType(t, "WaitGroup"):
		runtimeName = "GoCDRAllocWG"

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

	// TODO: what if only var m sync.Mutex
	case *ir.Decl:
		// var m sync.Mutex
		// if x.X == nil {
		// 	return nil, nil
		// }

		// return x.X.Type(), x.X
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

// ==================================================
// MARK: If
// ==================================================

func instrumentIfChain(n *ir.IfStmt) {
	numCases := countIfCases(n)

	caseNum := 0
	for cur := n; cur != nil; {

		cur.Body = instrumentStmtList(cur.Body)
		cur.Body = addControllRec(
			cur.Body,
			cur.Pos(),
			numCases,
			caseNum,
			"I",
		)

		caseNum++

		// else-if
		if len(cur.Else) == 1 {
			if next, ok := cur.Else[0].(*ir.IfStmt); ok {
				cur = next
				continue
			}
		}

		// final else
		if cur.Else != nil {
			cur.Else = instrumentStmtList(cur.Else)
			cur.Else = addControllRec(
				cur.Else,
				cur.Pos(),
				numCases,
				caseNum,
				"I",
			)
		}

		break
	}
}

func countIfCases(n *ir.IfStmt) int {
	count := 1

	for {
		if len(n.Else) == 1 {
			if next, ok := n.Else[0].(*ir.IfStmt); ok {
				count++
				n = next
				continue
			}
		}

		if n.Else != nil {
			count++
		}

		return count
	}
}

// ==================================================
// MARK: Switch
// ==================================================

func instrumentSwitch(n *ir.SwitchStmt) {
	numCases := len(n.Cases)

	for i, c := range n.Cases {
		c.Body = instrumentStmtList(c.Body)

		c.Body = addControllRec(
			c.Body,
			c.Pos(),
			numCases,
			i,
			"S",
		)
	}
}
func countSwitchCases(n *ir.SwitchStmt) int {
	count := 0

	for _, c := range n.Compiled {
		if _, ok := c.(*ir.CaseClause); ok {
			count++
		}
	}

	return count
}

// ==================================================
// MARK: Parameter
// ==================================================

func instrumentParameterCopy(fn *ir.Func) {
	calls := make(ir.Nodes, 0)
	for _, n := range fn.Dcl {
		if n.Class != ir.PPARAM {
			continue
		}

		// Skip *sync.Mutex.
		if n.Type().IsPtr() {
			continue
		}

		var runtimeName string

		switch {
		case isSyncType(n.Type(), "Mutex") || isSyncType(n.Type(), "RWMutex"):
			runtimeName = "GoCDRAllocMutex"

		case isSyncType(n.Type(), "Cond"):
			runtimeName = "GoCDRAllocCondVar"

		case isSyncType(n.Type(), "WaitGroup"):
			runtimeName = "GoCDRAllocWG"
		default:
			continue
		}

		f := typecheck.LookupRuntime(runtimeName)

		addr := ir.NewAddrExpr(n.Pos(), n)
		addr.SetType(types.NewPtr(n.Type()))
		addr.SetTypecheck(1)

		call := typecheck.Call(
			n.Pos(),
			f,
			[]ir.Node{makeUnsafePointer(addr, n.Pos())},
			false,
		)

		calls = append(calls, call)
	}

	if len(calls) > 0 {
		body := append(calls, fn.Body...)
		fn.Body = body
	}
}

// ==================================================
// MARK: shouldGoCDR
// ==================================================

func shouldGoCDR(n ir.Node) bool {
	if n == nil {
		return false
	}

	if !(base.Flag.GoCDRTrace || base.Flag.GoCDRReplay || base.Flag.GoCDRFuzzing) {
		return false
	}

	pkgName := n.Sym().Pkg.Path
	if pkgName == "runtime" ||
		pkgName == "syscall" ||
		pkgName == "os" ||
		pkgName == "fmt" ||
		strings.HasPrefix(pkgName, "internal") {

		return false
	}

	return !isGoCDRCall(n)
}

func isGoCDRCall(n ir.Node) bool {
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
		(fmt.Sprint(name.Sym()) == "GoCDRAllocMutex" ||
			fmt.Sprint(name.Sym()) == "GoCDRAllocCondVar" ||
			fmt.Sprint(name.Sym()) == "GoCDRAllocWG" ||
			fmt.Sprint(name.Sym()) == "gocdrTraceControllFlow")
}

// ==================================================
// MARK: helper
// ==================================================

func printFunc(fn *ir.Func) {
	fmt.Printf("FUNC: %v\n", fn.Sym())
	ir.DumpList("body", fn.Body)
}

func addControllRec(body ir.Nodes, pos src.XPos, numCases, caseNum int, t string) ir.Nodes {
	fn := typecheck.LookupRuntime("gocdrControllFlow")

	call := typecheck.Call(
		pos,
		fn,
		[]ir.Node{
			ir.NewString(pos, t),
			ir.NewInt(pos, int64(numCases)),
			ir.NewInt(pos, int64(caseNum)),
		},
		false,
	)

	out := make(ir.Nodes, 0, len(body)+1)
	out.Append(call)
	out.Append(body...)

	return out
}

func isUserMain(fn *ir.Func) bool {
	if fn == nil {
		return false
	}

	sym := fn.Sym()

	if sym.Name != "main" {
		return false
	}

	pkg := sym.Pkg
	if pkg == nil {
		return false
	}

	return pkg.Name == "main" && pkg.Path == "main"
}
