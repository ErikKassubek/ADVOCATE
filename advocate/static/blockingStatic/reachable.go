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
	"fmt"
	"go/ast"
)

// TODO: does not work if path needs to pass through in go func() {...}()
func (self *staticData) isReachableFuncFromFunc(start, target *ast.FuncDecl, calcPath bool) (bool, string) {
	if start == nil || target == nil {
		return false, ""
	}

	if start == target {
		return true, ""
	}

	type parentInfo struct {
		from *ast.FuncDecl
	}

	visited := make(map[*ast.FuncDecl]bool)
	parent := make(map[*ast.FuncDecl]parentInfo)

	queue := []*ast.FuncDecl{start}
	visited[start] = true

	found := false

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		info, ok := self.funcsInfo[cur]
		if !ok {
			continue
		}

		for _, n := range info.funcCalls {
			next := n.decl
			if next == nil {
				continue
			}

			if !visited[next] {
				visited[next] = true
				parent[next] = parentInfo{from: cur}

				if next == target {
					found = true
					queue = nil
					break
				}

				queue = append(queue, next)
			}
		}
	}

	if !found {
		return false, ""
	}

	if !calcPath {
		return true, ""
	}

	// Calculate the path from start to target.

	var path []*ast.FuncDecl

	for cur := target; cur != nil; cur = parent[cur].from {
		path = append(path, cur)
		if cur == start {
			break
		}
	}

	// reverse path
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	res := ""
	for i, fn := range path {
		if i > 0 {
			res += " -> "
		}
		res += fn.Name.Name
	}

	return true, res
}

// TODO: implement
func (self *staticData) isReachableRoutFromFunc(start *ast.FuncDecl, target *ast.GoStmt, calcPath bool) (bool, string) {
	targetFunc := self.routFunc[target]
	if targetFunc == nil { // TODO: make work with funcLit
		return false, ""
	}
	res, path := self.isReachableFuncFromFunc(start, targetFunc, calcPath)

	if !res || !calcPath {
		return res, ""
	}

	return res, fmt.Sprintf("%s -> Go", path)
}

// TODO: implement
func (self *staticData) isReachableFuncFromRout(start *ast.GoStmt, target *ast.FuncDecl, calcPath bool) (bool, string) {
	startFunc := self.routFunc[start]
	if startFunc == nil { // TODO: make work with funcLit
		return false, ""
	}

	res, path := self.isReachableFuncFromFunc(startFunc, target, calcPath)

	if !res || !calcPath {
		return res, ""
	}

	return res, fmt.Sprintf("Go -> %s", path)
}

// TODO: implement
func (self *staticData) isReachableObjFromFunc(start, target *ast.FuncDecl) bool {
	return false
}

// TODO: implement
func (self *staticData) isReachableObjFromRout(start, target *ast.FuncDecl) bool {
	return false
}
