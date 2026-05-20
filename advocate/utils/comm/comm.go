// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: comm.go
// Brief: Function to communicate between runtime and advocate
//
// Author: Erik Kassubek
// Created: 2026-04-20
//
// License: BSD-3-Clause

package comm

import (
	"advocate/utils/log"
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	OpenCom = true
	NoCom   = false
)

type CommProcess string

const (
	StaticBlock CommProcess = ":8080"
)

type Communication struct {
	process   CommProcess
	ln        net.Listener
	conn      net.Conn
	commMutex sync.Mutex
	isOpen    bool
}

func Open(process CommProcess) *Communication {
	comm := Communication{
		process:   process,
		commMutex: sync.Mutex{},
	}

	var err error
	comm.ln, err = net.Listen("tcp", string(process))
	if err != nil {
		log.Error("Communication Error: ", err.Error())
	}

	comm.conn, _ = comm.ln.Accept()
	comm.isOpen = true

	return &comm
}

func (self *Communication) Close() {
	self.commMutex.Lock()
	defer self.commMutex.Unlock()

	if !self.isOpen {
		return
	}

	self.conn.Close()
	self.ln.Close()
	self.isOpen = false
}

// Function to send a message to the runtime and return the response
//
// Parameter:
//   - msg string: message to send
//
// Returns:
//   - string: returned message
//   - error
func (self *Communication) Request(msg string) (string, error) {
	self.commMutex.Lock()
	defer self.commMutex.Unlock()
	if !self.isOpen {
		return "", fmt.Errorf("Comm not open")
	}
	self.Post(msg)
	return self.Get()
}

// Function to send a message to the runtime
//
// Parameter:
//   - msg string: message to send, must not contain the line "EOM"
//
// Returns:
//   - error
func (self *Communication) Post(msg string) error {
	self.commMutex.Lock()
	defer self.commMutex.Unlock()

	if !self.isOpen {
		return fmt.Errorf("Comm not open")
	}

	var err error
	for _, line := range strings.Split(msg, "\n") {
		_, err = fmt.Fprintln(self.conn, line)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(self.conn, "EOM")
	return err
}

// Function to recv a message from the runtime. End read if message is EOM
//
// Returns:
//   - string: recved message
//   - error
func (self *Communication) Get() (string, error) {
	self.commMutex.Lock()
	defer self.commMutex.Unlock()
	if !self.isOpen {
		return "", fmt.Errorf("Comm not open")
	}

	scanner := bufio.NewScanner(self.conn)

	res := ""

	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOM" {
			break
		}
		res += line + "\n"
	}

	if err := scanner.Err(); err != nil {
		return res, err
	}

	return strings.TrimRight(res, "\n"), nil
}
