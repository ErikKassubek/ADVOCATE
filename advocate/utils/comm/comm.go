// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: comm.go
// Brief: Function to communicate between runtime and advocate
//
// Author: Erik Kassubek, Mario Occhinegro
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

var ln net.Listener
var conn net.Conn
var commMutex sync.Mutex

var commIsOpen bool

func Open() {
	commMutex.Lock()
	defer commMutex.Unlock()

	if commIsOpen {
		return
	}

	log.Important("OPEN")
	var err error
	ln, err = net.Listen("tcp", ":8080")
	if err != nil {
		log.Error("Communication Error: ", err.Error())
	}

	conn, _ = ln.Accept()

	commIsOpen = true
}

func Close() {
	commMutex.Lock()
	defer commMutex.Unlock()

	if !commIsOpen {
		return
	}

	conn.Close()
	ln.Close()
	commIsOpen = false
}

// Function to send a message to the runtime and return the response
//
// Parameter:
//   - msg string: message to send
//
// Returns:
//   - string: returned message
//   - error
func Request(msg string) (string, error) {
	commMutex.Lock()
	if !commIsOpen {
		commMutex.Unlock()
		return "", fmt.Errorf("Comm not open")
	}
	commMutex.Unlock()
	Post(msg)
	return Get()
}

// Function to send a message to the runtime
//
// Parameter:
//   - msg string: message to send, must not contain the line "EOM"
//
// Returns:
//   - error
func Post(msg string) error {
	commMutex.Lock()
	defer commMutex.Unlock()
	if !commIsOpen {
		return fmt.Errorf("Comm not open")
	}

	var err error
	for _, line := range strings.Split(msg, "\n") {
		_, err = fmt.Fprintln(conn, line)
		if err != nil {
			log.Error(err)
		}
	}
	fmt.Fprintln(conn, "EOM")
	return err
}

// Function to recv a message from the runtime. End read if message is EOM
//
// Returns:
//   - string: recved message
//   - error
func Get() (string, error) {
	commMutex.Lock()
	defer commMutex.Unlock()
	if !commIsOpen {
		return "", fmt.Errorf("Comm not open")
	}

	scanner := bufio.NewScanner(conn)

	res := ""

	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOM" {
			log.Debug("BREAK")
			break
		}
		res += line + "\n"
	}

	if err := scanner.Err(); err != nil {
		return res, err
	}

	return strings.TrimRight(res, "\n"), nil
}
