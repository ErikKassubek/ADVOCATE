// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_comm.go
// Brief: Communication between runtime and advocate controller
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package advocatego

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

// Function to send a message to advocate and return the response
//
// Parameter:
//   - msg string: message to send
//
// Returns:
//   - string: returned message
//   - error
func AdvocateRequest(msg string) string {
	println("CONN")
	var conn, err = net.Dial("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}
	println("POST")
	AdvocatePost(conn, msg)
	println("GET")
	res := AdvocateGet(conn)
	println("RET")
	return res
}

// Function to send a message to advocate
//
// Parameter:"
//   - msg string: message to send, must not contain the line "EOM"
func AdvocatePost(conn io.Writer, msg string) {
	for _, line := range strings.Split(msg, "\n") {
		fmt.Fprintln(conn, line)
	}
	fmt.Fprintln(conn, "EOM")
}

// Function to recv a message to advocate. End read if message is EOM
//
// Returns:
//   - string: received message
//   - error
func AdvocateGet(conn io.Reader) string {
	scanner := bufio.NewScanner(conn)

	res := ""

	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOM" {
			break
		}
		res += line + "\n"
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("scanner error:", err)
	}

	return res
}
