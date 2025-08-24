# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestWGDoubleDone
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/WaitGroups/waitgroups_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/WaitGroups/waitgroups_test.go:18
```go
7 ...
8 
9 // ================================
10 // WaitGroup scenarios
11 // ================================
12 
13 // TestWGDoubleDone: wg.Done() called twice without Add
14 func TestWGDoubleDone(t *testing.T) {
15 	var wg sync.WaitGroup
16 	wg.Add(1)
17 	wg.Done()
18 	wg.Done() // Negative counter           // <-------
19 }
20 
21 // TestWGMissingDone: wg.Add() without corresponding Done
22 func TestWGMissingDone(t *testing.T) {
23 	var wg sync.WaitGroup
24 	wg.Add(1)
25 	go func() {
26 		// forgot wg.Done()
27 		time.Sleep(20 * time.Millisecond)
28 	}()
29 
30 ...
```


## Replay
**Replaying was not run**.

