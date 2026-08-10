package main

import (
	"fmt"
	"time"
)

// CODE CAN BE FOUND IN main_test.go

func main() {
}

func calc() int {
	time.Sleep(500 * time.Millisecond)
	return 1
}

func consume(x int) {
	fmt.Println(x)
}
