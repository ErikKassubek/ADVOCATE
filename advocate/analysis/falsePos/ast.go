// Copyright (c) 2025 Erik Kassubek
//
// File: ast.go
// Brief: Functions to work with the as
//
// Author: Erik Kassubek
// Created: 2025-11-17
//
// License: BSD-3-Clause

package falsepos

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

// parseFile creates an ast for a given file
//
// Parameter:
//   - filename string: the file name
//
// Returns:
//   - *token.FileSet: the ast file set
//   - *ast.File: the ast
//   - *types.Info: the type info
//   - error
func parseFile(fileName string) (*token.FileSet, *ast.File, *types.Info, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fileName, nil, 0)
	if err != nil {
		return nil, nil, nil, err
	}

	// Prepare type info
	conf := &types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	_, err = conf.Check("", fset, []*ast.File{file}, info)
	// Type errors are not fatal—we can still proceed
	if err != nil {
		fmt.Errorf("Warning: type checking not perfect:", err)
	}

	return fset, file, info, nil
}

// buildParentMap creates a map that, for each node in an ast contains the parent node
//
// Parameter:
//   - root *ast.File: the ast
//
// Returns:
//   - map[ast.Node]ast.Mode: maps each node to its parent in the ast
func buildParentMap(root ast.Node) map[ast.Node]ast.Node {
	parent := map[ast.Node]ast.Node{}

	var visit func(ast.Node)
	visit = func(n ast.Node) {
		ast.Inspect(n, func(child ast.Node) bool {
			if child != nil && child != n {
				parent[child] = n
			}
			return true
		})
	}

	visit(root)
	return parent
}

// findNodeAtLine searches for the node in the ast, that represents a specific line
//
// Parameter:
//   - fset *token.FileSet: the file set
//   - file *ast.File: the ast
//   - line int: the line to search for
//
// Returns:
//   - ast.Node: the ast Node that represents to operation in the given line
func findNodeAtLine(fset *token.FileSet, file *ast.File, line int) ast.Node {
	var found ast.Node

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		pos := fset.Position(n.Pos())
		if pos.Line == line {
			found = n
			return false
		}
		return true
	})

	return found
}

// finds the label attached to a loop: "<label>: for { ... }"
func findLabeledStmt(loop *ast.ForStmt, parentMap map[ast.Node]ast.Node) *ast.LabeledStmt {
	if parent, ok := parentMap[loop].(*ast.LabeledStmt); ok {
		return parent
	}
	return nil
}

// Checks if child belongs directly to the loop body (not a deeper nested loop)
func isDirectChildOf(parent, child ast.Node, parentMap map[ast.Node]ast.Node) bool {
	for n := child; n != nil; n = parentMap[n] {
		if n == parent {
			return true
		}
		// stop at other loop constructs
		switch n.(type) {
		case *ast.ForStmt, *ast.RangeStmt, *ast.SelectStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			if n != parent {
				return false
			}
		}
	}
	return false
}
