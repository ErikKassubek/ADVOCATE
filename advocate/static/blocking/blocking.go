package blocking

// var functionVar = make(map[string]map[string]map[string]struct{}{}) // function creation location -> variable -> function
// var funcInFunc = make(map[string][]string)                          // function creation location -> called created in function

// // varNames: variable names to look for
// // TODO: can this be made unique?
// func AreFunctionsAvailable(varNames []string) {
// 	fset := token.NewFileSet()

// 	parseFiles(fset)

// }

// func parseFiles(fset *token.FileSet, varNames []string) {
// 	files := make([]string, 0) // TODO: get files

// 	for _, file := range files {
// 		parseFile(fset, file)
// 	}
// }

// // for now only channel and mutex
// func parseFile(fset *token.FileSet, fileName string, varNames []string) error {
// 	file, err := parser.ParseFile(fset, fileName, nil, 0)
// 	if err != nil {
// 		panic(err)
// 	}

// 	funcname = file.Name.Name + ":" + fn.Name.Name // packageName:functionName

// 	for _, decl := range file.Decls {
// 		if fn, ok := decl.(*ast.FuncDecl); ok {
// 			// ast.Inspect(fn.Body, func(n ast.Node) bool) {

// 			// }
// 		}
// 	}
// }
