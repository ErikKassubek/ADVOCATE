# Leak: Leak on sync.WaitGroup

The analyzer detected a leak on a sync.WaitGroup.
A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.
A sync.WaitGroup wait is blocking, because the counter is not zero.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestWGDoubleDone
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/WaitGroups/waitgroups_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Waitgroup: Wait
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

