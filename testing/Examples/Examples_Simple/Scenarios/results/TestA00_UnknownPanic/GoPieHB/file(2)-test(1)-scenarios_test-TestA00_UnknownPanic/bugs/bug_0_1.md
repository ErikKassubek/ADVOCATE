# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestA00_UnknownPanic
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:12
```go
1 ...
2 
3 import (
4 	"sync"
5 	"testing"
6 	"time"
7 )
8 
9 // ---------------- A00–A07: Absolute Bugs ----------------
10 
11 func TestA00_UnknownPanic(t *testing.T) {
12 	panic("triggering unknown panic") // A00           // <-------
13 }
14 
15 func TestA01_SendOnClosed(t *testing.T) {
16 	ch := make(chan int)
17 	close(ch)
18 	ch <- 1 // A01
19 }
20 
21 func TestA02_ReceiveOnClosed(t *testing.T) {
22 	ch := make(chan int)
23 
24 ...
```


## Replay
**Replaying was not run**.

