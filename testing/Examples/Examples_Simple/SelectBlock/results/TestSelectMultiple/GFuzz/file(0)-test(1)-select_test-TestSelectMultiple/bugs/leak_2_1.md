# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestSelectMultiple
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/SelectBlock/select_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/SelectBlock/select_test.go:35
```go
24 ...
25 
26 	case ch <- 2: // partner exists
27 	case <-time.After(10 * time.Millisecond):
28 	}
29 }
30 
31 // TestSelectMultiple: multiple cases
32 func TestSelectMultiple(t *testing.T) {
33 	ch1 := make(chan int)
34 	ch2 := make(chan int)
35 	go func() { ch1 <- 1 }()           // <-------
36 	time.Sleep(5 * time.Millisecond)
37 	select {
38 	case <-ch1:
39 	case <-ch2:
40 	case <-time.After(20 * time.Millisecond):
41 	}
42 }
43 
```


## Replay
**Replaying was not run**.

