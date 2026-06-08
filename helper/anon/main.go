package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <path> <word>")
		os.Exit(1)
	}

	root := os.Args[1]
	a := os.Args[2]

	if err := process(root, a); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func process(root, a string) error {
	var dirsToRename []string
	var filesToRename []struct {
		oldPath string
		newPath string
	}
	var filesToRewrite []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()

		// Skip dot dirs
		if d.IsDir() && name != "." && strings.HasPrefix(name, ".") {
			return filepath.SkipDir
		}

		// collect dir rename
		if d.IsDir() && name == "advocate" {
			dirsToRename = append(dirsToRename, path)
			return nil
		}

		// delete executables called "advocate"
		if !d.IsDir() && name == "advocate" {
			info, err := d.Info()
			if err != nil {
				return err
			}

			// check executable bit (user/group/other)
			if info.Mode().IsRegular() && info.Mode()&0111 != 0 {
				if err := os.Remove(path); err != nil {
					return err
				}
				return nil
			}
		}

		// collect file rename
		if !d.IsDir() && strings.HasPrefix(name, "advocate_") {
			dir := filepath.Dir(path)
			newName := a + "_" + strings.TrimPrefix(name, "advocate_")
			newPath := filepath.Join(dir, newName)

			filesToRename = append(filesToRename, struct {
				oldPath string
				newPath string
			}{path, newPath})

			// IMPORTANT: do not rewrite original path after rename
			filesToRewrite = append(filesToRewrite, newPath)
			return nil
		}

		if !d.IsDir() {
			filesToRewrite = append(filesToRewrite, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// rename files
	for _, f := range filesToRename {
		if err := os.Rename(f.oldPath, f.newPath); err != nil {
			return err
		}
	}

	// rewrite files
	for _, path := range filesToRewrite {
		if err := rewriteFile(path, a); err != nil {
			return err
		}
	}

	// rename directories bottom-up
	for i := len(dirsToRename) - 1; i >= 0; i-- {
		oldPath := dirsToRename[i]
		newPath := filepath.Join(filepath.Dir(oldPath), a)

		if oldPath != newPath {
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}
	}

	// rename the root folder
	println(root, filepath.Dir(root))
	if err := os.Rename(root, filepath.Join(filepath.Dir(filepath.Clean(root)), capitalize(a))); err != nil {
		return err
	}

	return nil
}

func rewriteFile(path, a string) error {
	input, err := os.Open(path)
	if err != nil {
		return err
	}
	defer input.Close()

	reader := bufio.NewReader(input)

	var output strings.Builder

	for {
		line, err := reader.ReadString('\n')

		// EOF handling
		if err != nil && err.Error() != "EOF" {
			return err
		}

		trimmed := strings.TrimRight(line, "\n")

		// remove author line
		if strings.TrimSpace(trimmed) == "// Author: Erik Kassubek" {
			if err != nil {
				break
			}
			continue
		}

		// remove name occurrences
		trimmed = strings.ReplaceAll(trimmed, "Erik Kassubek", "")

		// replacements
		trimmed = strings.ReplaceAll(trimmed, "advocate", strings.ToLower(a))
		trimmed = strings.ReplaceAll(trimmed, "ADCOCATE", strings.ToUpper(a))
		trimmed = strings.ReplaceAll(trimmed, "Advocate", capitalize(a))

		output.WriteString(trimmed)

		if err == nil {
			output.WriteByte('\n')
		}

		if err != nil {
			break
		}
	}

	return os.WriteFile(path, []byte(output.String()), 0644)
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
}
