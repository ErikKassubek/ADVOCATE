// Copyright (c) 2026 Erik Kassubek
//
// File: reachable.go
// Brief: Determine if a func/routine/obj is reachable from a func/routine
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blockingStatic

import (
	"go/ast"
)

func (self *staticData) isReachableFuncFromFunc(start, target *ast.FuncDecl) bool {
	if start == nil || target == nil {
		return false
	}

	if start == target {
		return true
	}

	visited := make(map[*ast.FuncDecl]bool)

	queue := []*ast.FuncDecl{start}
	visited[start] = true

	for len(queue) > 0 {
		// pop front
		cur := queue[0]
		queue = queue[1:]

		// get outgoing edges
		funcs := self.funcsInfo[cur].funcCalls

		for _, n := range funcs {
			next := n.decl
			if next == nil {
				continue
			}

			if next == target {
				return true
			}

			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}

	return false
}

// TODO: implement
func (self *staticData) isReachableRoutFromFunc(start, target *ast.FuncDecl) bool {
	return false
}

// TODO: implement
func (self *staticData) isReachableObjFromFunc(start, target *ast.FuncDecl) bool {
	return false
}

// TODO: implement
func (self *staticData) isReachableFuncFromRout(start, target *ast.FuncDecl) bool {
	return false
}

// TODO: implement
func (self *staticData) isReachableObjFromRout(start, target *ast.FuncDecl) bool {
	return false
}
