# Explanation
This tool removes overhead from a a given file.
The preamble and import will be removed
# Input
# Output
# Usage
If a go file contains a main method the cool can be used like so
```sh
go run remover.go -f filename.go
```
# Example
Given a file `file.go`
```go
package main

import (
	"time"
    "advocate"
)

func main() {
	// ======= Preamble Start =======
		advocate.InitTracing(0)
		defer advocate.Finish()
	// ======= Preamble End =======
	c := make(chan int, 0)

	go func() {
		c <- 1
	}()

	go func() {
		<-c
	}()

	time.Sleep(10 * time.Millisecond)
	close(c)
}
```
After running `go run remover.go -f file.go` it will look like this
```go
package main

import (
	"time"
)

func main() {
	c := make(chan int, 0)

	go func() {
		c <- 1
	}()

	go func() {
		<-c
	}()

	time.Sleep(10 * time.Millisecond)
	close(c)
}
```