package experiments

import (
	"aplas/command"
	"aplas/data"
	"bytes"
	"fmt"
	"strconv"
	"time"
)

func sameElement() {
	wg.Add(1)
	defer wg.Done()

	outputName := "sameElementType"
	_, err := data.CreateOutput(outputName)
	if err != nil {
		fmt.Println(outputName + " -> " + err.Error())
	}

	for i := 0; i <= 1; i++ {
		go sameElementRun(outputName, i)
	}

}

func sameElementRun(outputName string, i int) {
	start()
	defer done()

	fmt.Printf("Start: %s %d\n", outputName, i)
	path, err := initialize(false, outputName, strconv.Itoa(i))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	settings := append(defaultSettings, []string{
		"-path",
		path,
		"-timeoutExec",
		"240",
		"-settings",
		fmt.Sprintf("SameElementTypeInOOC=%d", i),
	}...)

	start := time.Now()
	out := bytes.Buffer{}
	err = command.RunCommandAdvocate(&out, fmt.Sprintf("%s_%d", outputName, i), settings...)
	if err != nil {
		fmt.Println(err.Error())
	}
	dur := time.Since(start)

	writeData(outputName, strconv.Itoa(i), dur, out.String())

	fmt.Printf("Finish: %s %d\n", outputName, i)
}
