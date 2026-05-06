// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the static blocking analysis
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blockingStatic

import (
	"go/ast"
	"go/token"
)

// parse the files the determine the type information
func (self *staticData) collectOperations() {
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
		case *ast.GoStmt: // new routine
			self.recordGoStatement(fdecl, x)
		case *ast.SendStmt: // channel send
			self.recordOperation(fdecl, &x.Chan, chanSend)
		case *ast.UnaryExpr: // channel recv
			if x.Op == token.ARROW {
				self.recordOperation(fdecl, &x.X, chanRecv)
			}
		case *ast.CallExpr:
			// all functions
			self.recordFunctionCall(fdecl, x.Fun)

			// channel close
			if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "close" {
				self.recordOperation(fdecl, &(x.Args[0]), chanClose)
			}

			if sel, ok := x.Fun.(*ast.SelectorExpr); ok { // a.x()
				if self.isMutex(sel.Sel) {
					switch sel.Sel.Name {
					case "Lock":
						self.recordOperation(fdecl, &sel.X, mutexLock)
					case "TryLock":
						self.recordOperation(fdecl, &sel.X, mutexTryLock)
					case "RLock":
						self.recordOperation(fdecl, &sel.X, mutexRLock)
					case "TryRLock":
						self.recordOperation(fdecl, &sel.X, mutexRLock)
					case "Unlock":
						self.recordOperation(fdecl, &sel.X, mutexUnlock)
					}
				} else if self.isCondVar(sel.Sel) {
					switch sel.Sel.Name {
					case "Wait":
						self.recordOperation(fdecl, &sel.X, condWait)
					case "Signal":
						self.recordOperation(fdecl, &sel.X, condSignal)
					case "Broadcast":
						self.recordOperation(fdecl, &sel.X, condBroadcast)
					}
				} else if self.isWaitGroup(sel.Sel) {
					switch sel.Sel.Name {
					case "Wait":
						self.recordOperation(fdecl, &sel.X, wgWait)
					case "Add":
						self.recordOperation(fdecl, &sel.X, wgAdd)
					case "Done":
						self.recordOperation(fdecl, &sel.X, wgDone)
					case "Go":
						self.recordOperation(fdecl, &sel.X, wgGo)
					}
				}
			}
		}

		return true
	})
}

func (self *staticData) recordOperation(fdecl *ast.FuncDecl, variable *ast.Expr, f funcName) {
	if _, ok := self.opsPerFunk[fdecl]; !ok {
		self.opsPerFunk[fdecl] = make(map[*ast.Expr]map[funcName]struct{})
	}

	if _, ok := self.opsPerFunk[fdecl][variable]; !ok {
		self.opsPerFunk[fdecl][variable] = make(map[funcName]struct{})
	}

	self.opsPerFunk[fdecl][variable][f] = struct{}{}
}

func (self *staticData) recordFunctionCall(fdecl *ast.FuncDecl, f ast.Expr) {
	if _, ok := self.funcsPerFunc[fdecl]; !ok {
		self.funcsPerFunc[fdecl] = make([]ast.Expr, 0)
	}

	self.funcsPerFunc[fdecl] = append(self.funcsPerFunc[fdecl], f)
}

func (self *staticData) recordGoStatement(fdecl *ast.FuncDecl, f *ast.GoStmt) {
	if _, ok := self.funcsPerFunc[fdecl]; !ok {
		self.goStatementPerFunc[fdecl] = make([]*ast.GoStmt, 0)
	}

	self.goStatementPerFunc[fdecl] = append(self.goStatementPerFunc[fdecl], f)
}
