package main

import (
	"advocate/command"
	"advocate/experiments"
	"os"
	"os/signal"
	"syscall"
)

func cleanUp() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		command.Cancel()
		os.Exit(1)
	}()
}

func main() {
	cleanUp()
	// data.GetTestFuncs()
	experiments.Run()
}
