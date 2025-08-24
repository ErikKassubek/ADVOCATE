# Leak: Leak on sync.WaitGroup

The analyzer detected a leak on a sync.WaitGroup.
A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.
A sync.WaitGroup wait is blocking, because the counter is not zero.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixed3
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Waitgroup: Wait
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_mixed_test.go:53
```go
42 ...
43 
44 }
45 
46 // TestMixed3: P01 (possible send on closed), P03 (possible negative waitgroup), & L08 (mutex leak)
47 func TestMixed3(t *testing.T) {
48 	var wg sync.WaitGroup
49 	wg.Add(1)
50 	// Possible negative waitgroup inside goroutine
51 	go func() {
52 		wg.Done()
53 		wg.Done() // P03           // <-------
54 	}()
55 
56 	// Possible send on closed
57 	ch := make(chan int)
58 	go func() {
59 		close(ch)
60 	}()
61 	ch <- 10 // P01
62 
63 	// Leak on mutex: lock twice without unlock
64 
65 ...
```


## Replay
**Replaying was not run**.

