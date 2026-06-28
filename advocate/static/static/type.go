// Copyright (c) 2026 Erik Kassubek
//
// File: type.go
// Brief: Build type info
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package static

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

type funcName string

const (
	unknownFunc funcName = "<unknown>"
	makeFunc    funcName = "<make>"

	makeChan  funcName = "<chan:make>"
	chanSend  funcName = "<chan:send>"
	chanRecv  funcName = "<chan:recv>"
	chanClose funcName = "<chan:close>"

	mutexLock     funcName = "<mutex:lock>"
	mutexRLock    funcName = "<mutex:rlock>"
	mutexTryLock  funcName = "<mutex:trylock>"
	mutexTryRLock funcName = "<mutex:tryrlock>"
	mutexUnlock   funcName = "<mutex:unlock>"
	mutexRUnlock  funcName = "<mutex:runlock>"

	condWait      funcName = "<cond:wait>"
	condBroadcast funcName = "<cond:broadcast>"
	condSignal    funcName = "<cond:signal>"

	wgWait funcName = "<wait:wait>"
	wgAdd  funcName = "<wait:add>"
	wgDone funcName = "<wait:done>"
	wgGo   funcName = "<wait:go>"

	// TODO: list all
)

type objName int

const (
	unknownObj objName = iota

	mutex
	channel
	condVar
	wg
)

func (self *staticData) getName(id ast.Expr) string {
	if id == nil {
		return "NIL"
	}

	switch e := id.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return self.getName(e.X) + "." + e.Sel.Name
	case *ast.FuncLit:
		return "FuncLit"
	case *ast.CallExpr:
		return self.getName(e.Fun)
	default:
		panic(fmt.Sprintf("Unknown expr type %T", e))
	}
}

// parse the files the determine the type information
func (self *staticData) CollectOperations() {
	// per function
	for _, file := range self.ast {
		self.detOpsInFile(file)
	}
}

func (self *staticData) detOpsInFile(file *ast.File) {
	for _, decl := range file.Decls {
		fdecl, ok := decl.(*ast.FuncDecl)
		if !ok || fdecl.Body == nil {
			continue
		}
		self.getOpsInFunc(fdecl)
	}
}

func (self *staticData) getOpsInFunc(fdecl *ast.FuncDecl) {
	ast.Inspect(fdecl, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			self.funcDeclMap[x.Name.Pos()] = x
		case *ast.GoStmt: // new routine
			self.recordGoStatement(fdecl, x)
			self.recordFunctionCall(fdecl, x.Call)
		case *ast.SendStmt: // channel send
			self.recordOperation(fdecl, x.Chan, chanSend)
			return true
		case *ast.UnaryExpr: // channel recv
			if x.Op == token.ARROW {
				self.recordOperation(fdecl, x.X, chanRecv)
				return true
			}
		case *ast.CallExpr:
			ft := unknownFunc

			if ident, ok := x.Fun.(*ast.Ident); ok {
				switch ident.Name {
				case "close":
					if len(x.Args) == 1 && self.isChannel(x.Args[0]) {
						ft = chanClose
					}
				case "make":
					if len(x.Args) != 0 {
						switch x.Args[0].(type) {
						case *ast.ChanType:
							ft = makeChan
						}
					}
				}
				if ft != unknownFunc {
					self.recordOperation(fdecl, x.Args[0], ft)
					return true
				}
			}

			if sel, ok := x.Fun.(*ast.SelectorExpr); ok { // a.x()
				if self.isMutex(x) {
					switch sel.Sel.Name {
					case "Lock":
						ft = mutexLock
					case "TryLock":
						ft = mutexTryLock
					case "RLock":
						ft = mutexRLock
					case "TryRLock":
						ft = mutexRLock
					case "Unlock":
						ft = mutexUnlock
					case "RUnlock":
					}

				} else if self.isCondVar(x) {
					switch sel.Sel.Name {
					case "Wait":
						ft = condWait
					case "Signal":
						ft = condSignal
					case "Broadcast":
						ft = condBroadcast
					}
				} else if self.isWaitGroup(x) {

					switch sel.Sel.Name {
					case "Wait":
						ft = wgWait
					case "Add":
						ft = wgAdd
					case "Done":
						ft = wgDone
					case "Go":
						ft = wgGo
					}
				}

				if ft != unknownFunc {
					self.recordOperation(fdecl, sel.X, ft)
					return true
				}
			}
			self.recordFunctionCall(fdecl, x)
		}

		return true
	})
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
