// Copyright (c) 2026 Erik Kassubek
//
// File: reachable.go
// Brief: Determine if a func/routine/obj is reachable from a func/routine
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ast

import (
	"fmt"
	"go/ast"
)

// FuncFromFunc determines if the target function is reachable from the start function
// It is implemented as a BFS
//
// Parameter:
//   - start *ast.FuncDecl: function declaration to start from
//   - target *ast.FuncDecl: target function
//   - calcPath bool: if true, the path from start to traget is calculated. Otherwise it is only returned
//     whether a path exists
//
// Returns:
//   - bool: true if a path exists, false otherwise
//   - string: path from start to target. Only if path exists and calcPath is true
//   - error
func (self *Data) FuncFromFunc(start, target *ast.FuncDecl, calcPath bool) (bool, string, error) {
	if start == nil || target == nil {
		return false, "", fmt.Errorf("Start or target is nil")
	}

	if start == target {
		return true, "", nil
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

		info, ok := self.FuncInfo[cur]
		if !ok {
			continue
		}

		visit := func(next *ast.FuncDecl) bool {
			if next == nil {
				return false
			}

			if !visited[next] {
				visited[next] = true
				parent[next] = parentInfo{from: cur}

				if next == target {
					found = true
					queue = nil
					return true
				}

				queue = append(queue, next)
			}

			return false
		}

		for n := range info.FuncCalls {
			if visit(n.Decl) {
				break
			}
		}

		for _, n := range info.GoCalls {
			if visit(n) {
				break
			}
		}
	}

	if !found {
		return false, "", nil
	}

	if !calcPath {
		return true, "", nil
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
			res += " $ "
		}
		res += fn.Name.Name
	}

	return true, res, nil
}

func (self *Data) OpFromFunc(start *ast.FuncDecl, targetOp Operation, calcPath bool) (bool, string, error) {
	if start == nil {
		fmt.Println()
		return false, "", fmt.Errorf("Start is nil")
	}

	if self.FuncContainsOp(start, targetOp) {
		if calcPath {
			return true, start.Name.Name, nil
		}
		return true, "", nil
	}

	type parentInfo struct {
		from *ast.FuncDecl
	}

	visited := make(map[*ast.FuncDecl]bool)
	parent := make(map[*ast.FuncDecl]parentInfo)

	queue := []*ast.FuncDecl{start}
	visited[start] = true

	found := false

	var target *ast.FuncDecl

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		info, ok := self.FuncInfo[cur]
		if !ok {
			continue
		}

		visit := func(next *ast.FuncDecl) bool {
			if next == nil {
				return false
			}

			if !visited[next] {
				visited[next] = true
				parent[next] = parentInfo{from: cur}

				if self.FuncContainsOp(next, targetOp) {
					found = true
					queue = nil
					return true
				}

				queue = append(queue, next)
			}

			return false
		}

		for n := range info.FuncCalls {
			if visit(n.Decl) {
				target = n.Decl
				break
			}
		}

		for _, n := range info.GoCalls {
			if visit(n) {
				target = n
				break
			}
		}
	}

	if !found {
		return false, "", nil
	}

	if !calcPath {
		return true, "", nil
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
			res += " $ "
		}
		res += fn.Name.Name
	}
	res += " $ " + string(targetOp.Op)

	return true, res, nil
}
