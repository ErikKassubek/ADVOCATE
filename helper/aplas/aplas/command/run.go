package command

import (
	"aplas/data"
	"bytes"
	"context"
	"os/exec"
	"time"
)

var ctx, cancel = context.WithCancel(context.Background())

func Cancel() {
	cancel()
}

func RunCommandAdvocate(stdout *bytes.Buffer, args ...string) error {

	ctx, _ := context.WithTimeout(ctx, time.Duration(600)*time.Second)
	cmd := exec.CommandContext(ctx, data.Advocate, args...)
	cmd.Dir = data.PathAdvocate

	// cmd.Stdout = stdout
	cmd.Stderr = stdout

	return cmd.Run()
}

func RunCommandGo(stdout *bytes.Buffer, path string, args ...string) error {
	ctx, _ := context.WithTimeout(ctx, time.Duration(250)*time.Second)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = path

	// cmd.Stdout = stdout
	cmd.Stderr = stdout

	return cmd.Run()
}
