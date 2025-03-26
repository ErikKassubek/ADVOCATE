# Conditional
The three operations on conditional variables are
- Wait
- Signal
- Broadcast

## Trace element
The basic form of the trace element is
```
D,[tpre],[tpost],[id],[opN],[pos]
```
where `D` identifies the element as a conditional element.
The other fields are set as follows:
- [tpre] $\in \mathbb N$: This is the value of the global counter when the operation starts
the execution of the lock or unlock function
- [tpost] $\in \mathbb N$: This is the value of the global counter when the mutex has finished its operation. For lock operations this can be either if the lock was successfully acquired or if the routines continues its execution without
acquiring the lock in case of a trylock.
- [id] $\in \mathbb N$: This is the unique id identifying this mutex
- [opN]: This field shows the operation of the element. Those can be
  - [opM] = `W`: Wait
  - [opM] = `S`: Signal
  - [opM] = `B`: Broadcast
- [pos]: The last field show the position in the code, where the mutex operation
was executed. It consists of the file and line number separated by a colon (:)

## Example
 The following is an  example containing all the different recorded
elements:
```go
package main

import (
	"advocate"
	"fmt"
	"runtime"
	"sync"
	"time"
)

var (
	condition bool
	mu        sync.Mutex
	cond      *sync.Cond
)

func waitForCondition() {
	mu.Lock()
	defer mu.Unlock()

	for !condition {
		cond.Wait()
	}

	fmt.Println("Condition is true now!")
}

func signalCondition() {
	mu.Lock()
	defer mu.Unlock()

	condition = true
	cond.Signal()
}

func main() {
	cond = sync.NewCond(&mu)

	go waitForCondition()

	time.Sleep(time.Second) // Simulate some work

	signalCondition()
	time.Sleep(time.Second) // Allow time for the goroutine to print the message
}

```
If we ignore all unrelated internal operations, all atomic operations and assume that all operations are executed
before the program terminates, we get the following trace.
```txt
G,11,7,runtime.go:42;M,20,22,4,-,L,t,runtime.go:29;N,23,24,5,S,runtime.go:33;M,25,27,4,-,U,t,runtime.go:34;
```
```txt
M,12,14,4,-,L,t,runtime.go:18;N,15,31,5,W,runtime.go:22;M,17,19,4,-,U,t,/home/erik/Uni/HiWi/ADVOCATE/go-patch/src/sync/cond.go:83;M,28,30,4,-,L,t,/home/erik/Uni/HiWi/ADVOCATE/go-patch/src/sync/cond.go:85;M,32,34,6,-,L,t,/home/erik/Uni/HiWi/ADVOCATE/go-patch/src/sync/pool.go:216;M,36,38,6,-,U,t,/home/erik/Uni/HiWi/ADVOCATE/go-patch/src/sync/pool.go:233;M,51,53,4,-,U,t,runtime.go:26;
```

## Implementation
The recording of the mutex operations is implemented in the `go-patch/src/sync/cond.go` file in the implementation of the `Wait`, `Signal` und `Broadcast` functions.\
To save the id of the conditional, a field for the id is added to the `Cond` struct.\
The recording consist of two function calls, one at the beginning and one at the end of each function.
The first function call is called before the Operation tries to executed
and records the id ([id]) and called operation (opN), the position of the operation in the program ([pos]) and the counter at the beginning of the operation ([tpre]).\
The second function call records the success of the operation. This includes
the counter at the end of the operation ([tpost]).
The implementation of those function calls can be found in
`go-patch/src/runtime/advocate_trace.go` in the functions `AdvocateCondPre`, and `AdvocateCondPost`.
