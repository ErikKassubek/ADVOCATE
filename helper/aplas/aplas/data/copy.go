package data

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func CopyGobench(dst string) error {
	return CopyDir(PathGoBench, dst)
}

func CopyConstructed(dst string) error {
	return CopyDir(PathConstructed, dst)
}

func CopyDir(str, dst string) error {
	return filepath.WalkDir(str, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(str, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}
