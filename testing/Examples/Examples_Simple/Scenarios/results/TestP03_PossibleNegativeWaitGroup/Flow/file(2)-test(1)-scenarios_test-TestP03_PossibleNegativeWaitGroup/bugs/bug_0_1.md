# Bug: Actual negative Wait Group

During the execution, a negative waitgroup counter occured.
The occurrence of a negative wait group counter lead to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestP03_PossibleNegativeWaitGroup
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Waitgroup: Done
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:84
```go
73 ...
74 
75 	_ = <-ch // P02
76 	_ = <-ch // P02
77 }
78 
79 func TestP03_PossibleNegativeWaitGroup(t *testing.T) {
80 	var wg sync.WaitGroup
81 	wg.Add(1)
82 	go func() {
83 		wg.Done()
84 		wg.Done() // P03           // <-------
85 	}()
86 	wg.Wait()
87 }
88 
89 // ---------------- L00–L10: Leaks ----------------
90 
91 func TestL00_UnknownLeak(t *testing.T) {
92 	done := make(chan struct{})
93 	go func() { <-done }() // never signaled → L00
94 	time.Sleep(20 * time.Millisecond)
95 
96 ...
```


## Replay
**Replaying was not run**.

