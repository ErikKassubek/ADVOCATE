// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_comm.go
// Brief: Communication between runtime and advocate controller
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package advocate

import (
	"bufio"
	"fmt"
	"io"
	"net"
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
	var conn, err = net.Dial("tcp", "localhost:9000")

	if err != nil {
		panic("CON ERROR")
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
// Parameter:
//   - msg string: message to send
func AdvocatePost(conn io.Writer, msg string) {
	fmt.Fprintln(conn, msg)
	fmt.Fprintln(conn, "EOM")
}

// Function to recv a message to advocate. End read if message is EOM
//
// Returns:
//   - string: recved message
//   - error
func AdvocateGet(conn io.Reader) string {

	println("Create Scanner")
	scanner := bufio.NewScanner(conn)
	println("Created Scanner")

	res := ""

	println(scanner.Scan())
	for scanner.Scan() {
		line := scanner.Text()
		println("LINE: ", line)
		if line == "EOM" {
			println("EOM")
			break
		}
		res += line + "\n"
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("scanner error:", err)
	}

	return res
}
