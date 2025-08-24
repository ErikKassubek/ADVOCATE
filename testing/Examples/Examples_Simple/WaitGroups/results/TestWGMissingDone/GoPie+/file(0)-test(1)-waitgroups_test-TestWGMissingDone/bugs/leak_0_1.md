# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestWGMissingDone
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/WaitGroups/waitgroups_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/WaitGroups/waitgroups_test.go:29
```go
18 ...
19 
20 
21 // TestWGMissingDone: wg.Add() without corresponding Done
22 func TestWGMissingDone(t *testing.T) {
23 	var wg sync.WaitGroup
24 	wg.Add(1)
25 	go func() {
26 		// forgot wg.Done()
27 		time.Sleep(20 * time.Millisecond)
28 	}()
29 	wg.Wait() // Leak           // <-------
30 }
31 
32 // TestWGNested: nested adds and dones
33 func TestWGNested(t *testing.T) {
34 	var wg sync.WaitGroup
35 	wg.Add(2)
36 	go func() {
37 		wg.Done()
38 		go func() {
39 			wg.Done()
40 
41 ...
```


## Replay
**Replaying was not run**.

