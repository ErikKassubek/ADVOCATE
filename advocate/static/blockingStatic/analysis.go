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
