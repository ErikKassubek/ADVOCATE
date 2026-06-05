package experiments

import (
	"aplas/command"
	"aplas/data"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func realWorld() {
	wg.Add(1)
	defer wg.Done()

	names := getProg()

	for _, name := range names {
		go runRealWorld(name, filepath.Join(data.PathRealWorld, name))
	}
}

func runRealWorld(name, path string) {
	start()
	defer done()

	rwBase(name, path)
	rwFull(name, path)
}

func getProg() []string {
	entries, err := os.ReadDir(data.PathRealWorld)
	if err != nil {
		println(err.Error())
		return nil
	}

	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(entry.Name()))
		}
	}

	return dirs
}

func rwBase(name, path string) {
	wg.Add(1)
	defer wg.Done()

	fmt.Printf("Start: RW base\n")

	start := time.Now()
	out := bytes.Buffer{}
	absPath, _ := filepath.Abs(path)
	err := command.RunCommandGo(&out, fmt.Sprintf("base_rw_%s", name), absPath, "test", "./...", "-count=1", "timeout", "240s")
	if err != nil {
		println(out.String())
		fmt.Println(err.Error())
	}
	dur := time.Since(start)

	writeData("rwBase", name, dur, "")
}

func rwFull(name, path string) {
	start()
	defer done()

	start := time.Now()
	out := bytes.Buffer{}

	settings := append(defaultSettings, []string{
		"-path",
		path,
		"-timeoutExec",
		"240",
	}...)

	err := command.RunCommandAdvocate(&out, fmt.Sprintf("rw_%s", name), settings...)
	if err != nil {
		fmt.Println(err.Error())
	}
	dur := time.Since(start)

	writeData("rw", name, dur, "")
}
