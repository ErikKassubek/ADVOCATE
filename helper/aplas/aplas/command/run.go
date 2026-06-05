package command

import (
	"aplas/data"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var ctx, cancel = context.WithCancel(context.Background())

func Cancel() {
	cancel()
}

func RunCommandAdvocate(stdout *bytes.Buffer, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(6)*time.Hour)
	cmd := exec.CommandContext(ctx, data.Advocate, args...)
	cmd.Dir = data.PathAdvocate

	logFile := filepath.Join(data.PathLog, "log"+"_"+name+".log")

	f, err := os.Create(logFile)
	if err != nil {
		cancel()
		return err
	}
	defer f.Close()

	mw := io.MultiWriter(stdout, f)

	cmd.Stdout = mw
	cmd.Stderr = mw

	err = cmd.Run()

	cancel()

	return err
}

func RunCommandGo(stdout *bytes.Buffer, name string, path string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	logFile := filepath.Join(data.PathLog, "log"+"_"+name+".log")

	f, err := os.Create(logFile)
	if err != nil {
		cancel()
		return err
	}
	defer f.Close()

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = path

	mw := io.MultiWriter(stdout, f)

	cmd.Stdout = mw
	cmd.Stderr = mw

	err = cmd.Run()

	cancel()

	return err
}
