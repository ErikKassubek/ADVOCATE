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

	log.Debug("OPEN")
	var err error
	ln, err = net.Listen("tcp", ":9000")
	if err != nil {
		log.Debug("ERROR: ", err.Error())
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

	log.Debug("CLOSE")
	conn.Close()
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
//   - msg string: message to send
//
// Returns:
//   - error
func Post(msg string) error {
	log.Debug("POST")
	commMutex.Lock()
	log.Debug("POST 0")
	defer commMutex.Unlock()
	if !commIsOpen {
		return fmt.Errorf("Comm not open")
	}

	log.Debug("Post 1")
	_, err := fmt.Fprintln(conn, msg)
	log.Debug("Post 2")
	fmt.Fprintln(conn, "EOM")
	log.Debug("Post 3")
	return err
}

// Function to recv a message from the runtime. End read if message is EOM
//
// Returns:
//   - string: recved message
//   - error
func Get() (string, error) {
	log.Debug("GET")
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
