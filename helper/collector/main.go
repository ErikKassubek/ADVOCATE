package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
)

func main() {
	var src string
	var dest string

	flag.StringVar(&src, "src", "", "Source")
	flag.StringVar(&dest, "dest", "", "Destination")

	flag.Parse()

	if src == "" {
		panic("Required element -src [arg] missing")
	}

	if dest == "" {
		panic("Required element -dest [arg] missing")
	}

	progs, err := os.ReadDir(src)
	if err != nil {
		panic(err)
	}

	for _, prog := range progs {
		resPath := filepath.Join(src, prog.Name(), "advocateResult")

		tests, err := os.ReadDir(resPath)
		if err != nil {
			continue
		}
		for _, test := range tests {
			bugPath := filepath.Join(resPath, test.Name(), "bugs")
			if _, err := os.Stat(bugPath); os.IsNotExist(err) {
				continue
			}
			copyDir(bugPath, dest)
		}
	}

}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	// parent folder name (e.g. "b" from "/a/b/c")
	parentName := filepath.Base(filepath.Dir(src))

	// destination root becomes dst/<parentName>
	targetRoot := filepath.Join(dst, parentName)

	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetRoot, rel)

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath, info.Mode())
	})
}
