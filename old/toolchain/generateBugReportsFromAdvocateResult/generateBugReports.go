package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	folderName := flag.String("f", "", "path to the folder")
	advocateroot := flag.String("a", "", "path to the advocate root")
	flag.Parse()
	if *folderName == "" || *advocateroot == "" {
		fmt.Fprintln(os.Stderr, "Usage generateBugReports -f <folder> -a <advocate root>")
		os.Exit(1)
	}
	files, err := getFiles(*folderName, "results_machine.log")
	if err != nil {
		fmt.Println(err)
	}
	analyzerPath := filepath.Join(*advocateroot, "analyzer", "analyzer")
	for _, file := range files {
		folder := filepath.Dir(file)
		advocateTraceFolder := filepath.Join(folder, "advocateTrace")
		cmd := exec.Command("wc", "-l", file)
		out, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
		}
		lineCount, err := strconv.Atoi(strings.Fields(string(out))[0])
		if err != nil {
			fmt.Println(err)
		}
		for i := 1; i <= lineCount; i++ {
			cmd := exec.Command(analyzerPath, "-e", "-trace", advocateTraceFolder, "-i", strconv.Itoa(i))
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
		}

	}
}

func getFiles(folderPath string, fileName string) ([]string, error) {
	var files []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == fileName {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
