package data

import (
	"io/fs"
	"path/filepath"
	"strings"
)

var tests []string

func fileToTest(path string) string {
	file := strings.Split(filepath.Base(path), "_")[0]
	return "Test" + strings.ToUpper(file[:1]) + file[1:]
}

func GetTestFuncs() []string {
	if len(tests) != 0 {
		return tests
	}

	tests = make([]string, 0)
	filepath.WalkDir(PathGoBench, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			tests = append(tests, fileToTest(path))
		}

		return nil
	})

	return tests
}
