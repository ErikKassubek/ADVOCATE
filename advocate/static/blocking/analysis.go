// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the static blocking analysis
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blocking

import (
	"go/ast"
	"go/token"
)

// parse the files the determine the type information
func (self *staticData) determineOperations() {

	for _, file := range self.ast {
		self.getOpsInFile(file)
	}
}

func (self *staticData) getOpsInFile(file *ast.File) {
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
		// channel send
		case *ast.SendStmt:
			self.recordOperation(fdecl, &x.Chan, chanSend)
		case *ast.UnaryExpr:
			// channel recv
			if x.Op == token.ARROW {
				self.recordOperation(fdecl, &x.X, chanRecv)
			}

		case *ast.CallExpr:
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
				}

				// TODO: continue
			}
		}

		return true
	})
}

func (self *staticData) recordOperation(fdecl *ast.FuncDecl, variable *ast.Expr, f funcs) {
	if _, ok := self.opsPerFunk[fdecl]; !ok {
		self.opsPerFunk[fdecl] = make(map[*ast.Expr]map[funcs]struct{})
	}

	if _, ok := self.opsPerFunk[fdecl][variable]; !ok {
		self.opsPerFunk[fdecl][variable] = make(map[funcs]struct{})
	}

	self.opsPerFunk[fdecl][variable][f] = struct{}{}
}

func (self *staticData) isMutex(id *ast.Ident) bool {
	named, ok := self.getNamed(id)

	return ok && named.Obj().Pkg().Path() == "sync" &&
		(named.Obj().Name() == "Mutex" || named.Obj().Name() == "RWMutex")
}
