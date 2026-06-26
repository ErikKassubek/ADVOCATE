// Copyright (c) 2026 Erik Kassubek
//
// File: type.go
// Brief: Build type info
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blockingStatic

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func (self *staticData) buildTypeInfo() {
	self.pkgInfo = make(map[*packages.Package]*types.Info)

	self.uses = make(map[*ast.Ident]types.Object)
	self.defs = make(map[*ast.Ident]types.Object)

	for _, pkg := range self.pkgs {
		if pkg.TypesInfo == nil {
			continue
		}

		self.pkgInfo[pkg] = pkg.TypesInfo

		for ident, obj := range pkg.TypesInfo.Uses {
			self.uses[ident] = obj
		}

		for ident, obj := range pkg.TypesInfo.Defs {
			self.defs[ident] = obj
		}
	}
}

func (self *staticData) isConcOp(id *ast.CallExpr) string {
	if self.isMutex(id) {
		return "Mutex"
	} else if self.isCondVar(id) {
		return "CondVar"
	} else if self.isWaitGroup(id) {
		return "WaitGroup"
	} else {
		return "?"
	}
}

func (self *staticData) isMutex(id *ast.CallExpr) bool {
	return self.isConcObj(id, mutex)
}

func (self *staticData) isCondVar(id *ast.CallExpr) bool {
	return self.isConcObj(id, condVar)
}

func (self *staticData) isWaitGroup(id *ast.CallExpr) bool {
	return self.isConcObj(id, wg)
}

// TODO: channel
func (self *staticData) isConcObj(call *ast.CallExpr, on objName) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)

	// Find the package containing this node.
	pkg := self.npm[call]
	if pkg == nil {
		println("Not Sel")
		return false
	}

	info := self.pkgInfo[pkg]
	if info == nil {
		return false
	}

	selection := info.Selections[sel]
	if selection == nil {
		// Not a method call (e.g. pkg.Func()).
		return false
	}

	t := selection.Recv()

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	switch on {
	case mutex:
		return obj.Name() == "Mutex" &&
			obj.Pkg().Path() == "sync"
	case wg:
		return obj.Name() == "WaitGroup" &&
			obj.Pkg().Path() == "sync"
	case condVar:
		return obj.Name() == "CondVar" &&
			obj.Pkg().Path() == "sync"
	}

	return false
}

func (self *staticData) isChannel(expr ast.Expr) bool {
	// Find the package containing this node.
	pkg := self.npm[expr]
	if pkg == nil {
		return false
	}

	info := self.pkgInfo[pkg]
	if info == nil {
		return false
	}

	tv, ok := info.Types[expr]
	if !ok || tv.Type == nil {
		return false
	}

	_, ok = tv.Type.Underlying().(*types.Chan)
	return ok
}

// TODO: channel send
func (self *staticData) getConcFuncName(id ast.Expr) funcName {
	// if x, ok := id.(*ast.SelectorExpr); ok { // X.Sel is func name
	// 	// fmt.Println(self.isMutex(x.Fun))
	// 	// } else {
	// 	// 	fmt.Println("Not Ident")
	// 	fmt.Println("is call ", x.Sel, self.isMutex(x.Sel))
	// } else {
	// 	fmt.Println("is not")
	// }
	switch x := id.(type) {
	// case *ast.SendStmt: // channel send
	// 	return "<chan:send>"
	case *ast.UnaryExpr: // channel recv
		if x.Op == token.ARROW {
			return chanRecv
		}
	case *ast.CallExpr:
		// channel close
		if ident, ok := x.Fun.(*ast.Ident); ok {
			switch ident.Name {
			case "close":
				if len(x.Args) != 1 {
					return unknownFunc
				}
				if self.isChannel(x.Args[0]) {
					return chanClose
				}
				return unknownFunc
			case "make":
				if len(x.Args) == 0 {
					return unknownFunc
				}

				switch x.Args[0].(type) {
				case *ast.ChanType:
					return makeChan
				}
				return makeFunc
			}
		}

		if sel, ok := x.Fun.(*ast.SelectorExpr); ok { // a.x()
			if self.isMutex(x) {
				switch sel.Sel.Name {
				case "Lock":
					return mutexLock
				case "TryLock":
					return mutexTryLock
				case "RLock":
					return mutexRLock
				case "TryRLock":
					return mutexTryRLock
				case "Unlock":
					return mutexUnlock
				case "RUnlock":
					return mutexRUnlock
				}
			} else if self.isCondVar(x) {
				switch sel.Sel.Name {
				case "Wait":
					return condWait
				case "Signal":
					return condSignal
				case "Broadcast":
					return condBroadcast
				}
			} else if self.isWaitGroup(x) {
				switch sel.Sel.Name {
				case "Wait":
					return wgWait
				case "Add":
					return wgAdd
				case "Done":
					return wgDone
				case "Go":
					return wgGo
				}
			}
		}
	}

	return unknownFunc
}
