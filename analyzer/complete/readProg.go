// Copyrigth (c) 2024 Erik Kassubek
//
// File: readProg.go
// Brief: Functions to read in a program an extrace all relevant operations
//
// Author: Erik Kassubek
// Created: 2024-06-26
//
// License: BSD-3-Clause

package complete

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getProgramElements(progPath string) (map[string][]int, error) {
	progElems := make(map[string][]int)

	files, err := collectGoFiles(progPath)
	if err != nil {
		log.Println("Error in collecting files")
		return nil, err
	}

	pkg, err := analyzeFiles(files)
	if err != nil {
		log.Println("Error in analyzing files")
		return nil, err
	}

	// traverse all .go files in the directory recursively
	err = filepath.Walk(progPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error in walking")
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			content, err := os.ReadFile(path)
			if err != nil {
				log.Println("Error in reading file")
				return err
			}

			elems, err := getElemsFromContent(path, string(content), pkg)

			if err == nil && len(elems) > 0 {
				if _, ok := progElems[path]; !ok {
					progElems[path] = make([]int, 0)
				}
				for _, elem := range elems {
					progElems[path] = append(progElems[path], elem)
				}
			}
		}
		return nil
	})

	return progElems, err
}

func collectGoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error when collecting %q: %v\n", path, err)
			return err
		}
		if info == nil {
			return nil
		}

		if info.IsDir() && info.Name() == "advocateResult" {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func analyzeFiles(files []string) (*types.Package, error) {
	fset := token.NewFileSet()
	var astFiles []*ast.File

	for _, file := range files {
		parsedFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		astFiles = append(astFiles, parsedFile)
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, _ := conf.Check("mypackage", fset, astFiles, &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	})

	return pkg, nil
}

func getElemsFromContent(path string, content string, pkg *types.Package) ([]int, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return []int{}, err
	}

	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}

	imports := pkg.Imports()

	var syncPkg *types.Package
	for _, imp := range imports {
		if imp.Name() == "sync" {
			syncPkg = imp
			break
		}
	}

	v := &visitor{fset: fset, pkg: pkg, info: info, syncPkg: syncPkg,
		selectCases: make(map[string]struct{}), elements: make([]int, 0)}
	ast.Walk(v, node)

	return v.elements, nil
}

// visitor ist eine Struktur, die das ast.Visitor Interface implementiert.
// Sie wird verwendet, um den AST zu durchlaufen.
type visitor struct {
	fset        *token.FileSet
	pkg         *types.Package
	info        *types.Info
	syncPkg     *types.Package
	selectCases map[string]struct{}
	elements    []int // line numbers
}

// Visit wird für jeden Knoten im AST aufgerufen.
func (v *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch x := n.(type) {
	case *ast.GoStmt:
		v.recordElement(v.fset.Position(n.Pos()))
	case *ast.SendStmt: // send
		if _, ok := v.selectCases[v.fset.Position(n.Pos()).String()]; ok {
			delete(v.selectCases, v.fset.Position(n.Pos()).String())
		} else {
			v.recordElement(v.fset.Position(n.Pos()))
		}
	case *ast.UnaryExpr: // recv
		if x.Op == token.ARROW {
			if _, ok := v.selectCases[v.fset.Position(n.Pos()).String()]; ok {
				delete(v.selectCases, v.fset.Position(n.Pos()).String())
			} else {
				v.recordElement(v.fset.Position(n.Pos()))
			}
		}
	case *ast.CallExpr:
		// close
		if fun, ok := x.Fun.(*ast.Ident); ok && fun.Name == "close" {
			v.recordElement(v.fset.Position(n.Pos()))
		}
		if fun, ok := x.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := fun.X.(*ast.Ident); ok {
				obj := v.info.Uses[ident]

				if obj != nil && v.syncPkg != nil {
					typ := obj.Type()

					// Überprüfen Sie, ob der Typ zu einem der spezifischen Typen gehört
					mutexType := v.syncPkg.Scope().Lookup("Mutex").Type()
					rwMutexType := v.syncPkg.Scope().Lookup("RWMutex").Type()
					wgType := v.syncPkg.Scope().Lookup("WaitGroup").Type()
					condType := v.syncPkg.Scope().Lookup("Cond").Type()
					onceType := v.syncPkg.Scope().Lookup("Once").Type()

					switch {
					case types.AssignableTo(typ, mutexType):
						v.recordElement(v.fset.Position(n.Pos()))
					case types.AssignableTo(typ, rwMutexType):
						v.recordElement(v.fset.Position(n.Pos()))
					case types.AssignableTo(typ, wgType):
						v.recordElement(v.fset.Position(n.Pos()))
					case types.AssignableTo(typ, condType):
						v.recordElement(v.fset.Position(n.Pos()))
					case types.AssignableTo(typ, onceType):
						v.recordElement(v.fset.Position(n.Pos()))
					}
				}
			}
		}
	case *ast.SelectStmt:
		v.recordElement(v.fset.Position(n.Pos()))
		for _, stmt := range x.Body.List {
			caseClause, ok := stmt.(*ast.CommClause)
			if !ok {
				continue // Nicht ein case-timeouteil, weitermachen
			}
			switch comm := caseClause.Comm.(type) {
			case *ast.SendStmt:
				// store to not record the send statement
				v.selectCases[v.fset.Position(comm.Pos()).String()] = struct{}{}
			case *ast.ExprStmt:
				// store to not record the recv statement
				if unaryExpr, ok := comm.X.(*ast.UnaryExpr); ok && unaryExpr.Op == token.ARROW {
					v.selectCases[v.fset.Position(unaryExpr.Pos()).String()] = struct{}{}
				}
			}
		}
	case *ast.RangeStmt:
		rangeExpr := x.X
		rangeExprType := v.info.Types[rangeExpr].Type
		// Check if the range expression is a channel
		if _, ok := rangeExprType.(*types.Chan); ok {
			fmt.Printf("Range über Kanal gefunden bei %s\n", v.fset.Position(n.Pos()))
		}
	}

	return v
}

func (v *visitor) recordElement(pos token.Position) {
	v.elements = append(v.elements, pos.Line)
}
