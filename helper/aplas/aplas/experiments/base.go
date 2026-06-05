package experiments

import (
	"aplas/command"
	"bytes"
	"fmt"
	"path/filepath"
	"time"
)

func base() {
	wg.Add(1)
	go func() {
		defer wg.Done()

		start()
		defer done()

		fmt.Printf("Start: Base\n")

		outputName := "base"
		path, err := initialize(false, outputName)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		start := time.Now()
		out := bytes.Buffer{}
		absPath, _ := filepath.Abs(path)
		err = command.RunCommandGo(&out, absPath, "test", "./...", "-count=1", "timeout", "20s")
		if err != nil {
			fmt.Println(err.Error())
		}
		dur := time.Since(start)

		writeData(outputName, "Base", dur, out.String())
	}()

}
