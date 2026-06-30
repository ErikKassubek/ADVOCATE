// Copyright (c) 2026 Erik Kassubek
//
// File: type.go
// Brief: Build type info
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package staticAST

import (
	"advocate/static/static/staticBase"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

func (self *Data) getName(id ast.Expr) string {
	if id == nil {
		return "NIL"
	}

	return self.getPackage(id) + ":" + self.getNameRec(id)
}

func (self *Data) getNameRec(id ast.Expr) string {
	switch e := id.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return self.getNameRec(e.X) + "." + e.Sel.Name
	case *ast.FuncLit:
		return "FuncLit"
	case *ast.CallExpr:
		return self.getNameRec(e.Fun)
	default:
		panic(fmt.Sprintf("Unknown expr type %T", e))
	}
}

func (self *Data) getPackage(expr ast.Expr) string {
	pkg, ok := self.npm[expr]
	if !ok {
		return "[unknown]"
	}
	return pkg.Name
}

// parse the files the determine the type information
func (self *Data) CollectOperations() {
	// per function
	for _, file := range self.ast {
		self.detOpsInFile(file)
	}
}

func (self *Data) detOpsInFile(file *ast.File) {
	for _, decl := range file.Decls {
		fdecl, ok := decl.(*ast.FuncDecl)
		if !ok || fdecl.Body == nil {
			continue
		}
		self.detOpsInFunc(fdecl)
	}
}

func (self *Data) detOpsInFunc(fdecl *ast.FuncDecl) {
	self.addFuncIfNotExists(fdecl)

	ast.Inspect(fdecl, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			self.funcDeclMap[x.Name.Pos()] = x
		case *ast.GoStmt: // new routine
			self.recordGoStatement(fdecl, x)
			// self.recordFunctionCall(fdecl, x.Call)
		case *ast.SendStmt: // channel send
			self.recordOperation(fdecl, x.Chan, staticBase.ChanSend)
			return true
		case *ast.UnaryExpr: // channel recv
			if x.Op == token.ARROW {
				self.recordOperation(fdecl, x.X, staticBase.ChanRecv)
				return true
			}
		case *ast.CallExpr:
			ft := staticBase.UnknownFunc

			if ident, ok := x.Fun.(*ast.Ident); ok {
				switch ident.Name {
				case "close":
					if len(x.Args) == 1 && self.isChannel(x.Args[0]) {
						ft = staticBase.ChanClose
					}
				case "make":
					if len(x.Args) != 0 {
						switch x.Args[0].(type) {
						case *ast.ChanType:
							ft = staticBase.MakeChan
						}
					}
				}
				if ft != staticBase.UnknownFunc {
					self.recordOperation(fdecl, x.Args[0], ft)
					return true
				}
			}

			if sel, ok := x.Fun.(*ast.SelectorExpr); ok { // a.x()
				if self.isMutex(x) {
					switch sel.Sel.Name {
					case "Lock":
						ft = staticBase.MutexLock
					case "TryLock":
						ft = staticBase.MutexTryLock
					case "RLock":
						ft = staticBase.MutexRLock
					case "TryRLock":
						ft = staticBase.MutexRLock
					case "Unlock":
						ft = staticBase.MutexUnlock
					case "RUnlock":
						ft = staticBase.MutexRUnlock
					}

				} else if self.isCondVar(x) {
					switch sel.Sel.Name {
					case "Wait":
						ft = staticBase.CondWait
					case "Signal":
						ft = staticBase.CondSignal
					case "Broadcast":
						ft = staticBase.CondBroadcast
					}
				} else if self.isWaitGroup(x) {

					switch sel.Sel.Name {
					case "Wait":
						ft = staticBase.WgWait
					case "Add":
						ft = staticBase.WgAdd
					case "Done":
						ft = staticBase.WgDone
					case "Go":
						ft = staticBase.WgGo
					}
				}

				if ft != staticBase.UnknownFunc {
					self.recordOperation(fdecl, sel.X, ft)
					return true
				}
			}
			self.recordFunctionCall(fdecl, x)
		}

		return true
	})
}

func (self *Data) isConcOp(id *ast.CallExpr) string {
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

func (self *Data) isMutex(id *ast.CallExpr) bool {
	return self.isConcObj(id, staticBase.Mutex)
}

func (self *Data) isCondVar(id *ast.CallExpr) bool {
	return self.isConcObj(id, staticBase.CondVar)
}

func (self *Data) isWaitGroup(id *ast.CallExpr) bool {
	return self.isConcObj(id, staticBase.Wg)
}

// TODO: channel
func (self *Data) isConcObj(call *ast.CallExpr, on staticBase.ObjName) bool {
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
	case staticBase.Mutex:
		return (obj.Name() == "Mutex" || obj.Name() == "RWMutex") &&
			obj.Pkg().Path() == "sync"
	case staticBase.Wg:
		return obj.Name() == "WaitGroup" &&
			obj.Pkg().Path() == "sync"
	case staticBase.CondVar:
		return obj.Name() == "CondVar" &&
			obj.Pkg().Path() == "sync"
	}

	return false
}

func (self *Data) isChannel(expr ast.Expr) bool {
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
func (self *Data) getConcFuncName(id ast.Expr) staticBase.FuncName {
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
			return staticBase.ChanRecv
		}
	case *ast.CallExpr:
		// channel close
		if ident, ok := x.Fun.(*ast.Ident); ok {
			switch ident.Name {
			case "close":
				if len(x.Args) != 1 {
					return staticBase.UnknownFunc
				}
				if self.isChannel(x.Args[0]) {
					return staticBase.ChanClose
				}
				return staticBase.UnknownFunc
			case "make":
				if len(x.Args) == 0 {
					return staticBase.UnknownFunc
				}

				switch x.Args[0].(type) {
				case *ast.ChanType:
					return staticBase.MakeChan
				}
				return staticBase.MakeFunc
			}
		}

		if sel, ok := x.Fun.(*ast.SelectorExpr); ok { // a.x()
			if self.isMutex(x) {
				switch sel.Sel.Name {
				case "Lock":
					return staticBase.MutexLock
				case "TryLock":
					return staticBase.MutexTryLock
				case "RLock":
					return staticBase.MutexRLock
				case "TryRLock":
					return staticBase.MutexTryRLock
				case "Unlock":
					return staticBase.MutexUnlock
				case "RUnlock":
					return staticBase.MutexRUnlock
				}
			} else if self.isCondVar(x) {
				switch sel.Sel.Name {
				case "Wait":
					return staticBase.CondWait
				case "Signal":
					return staticBase.CondSignal
				case "Broadcast":
					return staticBase.CondBroadcast
				}
			} else if self.isWaitGroup(x) {
				switch sel.Sel.Name {
				case "Wait":
					return staticBase.WgWait
				case "Add":
					return staticBase.WgAdd
				case "Done":
					return staticBase.WgDone
				case "Go":
					return staticBase.WgGo
				}
			}
		}
	}

	return staticBase.UnknownFunc
}
