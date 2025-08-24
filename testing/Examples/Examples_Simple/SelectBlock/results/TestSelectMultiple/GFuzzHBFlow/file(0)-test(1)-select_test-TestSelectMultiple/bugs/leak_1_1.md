# Leak: Leak on unbuffered channel with possible partner

The analyzer detected a Leak on an unbuffered channel with a possible partner.
A Leak on an unbuffered channel is a situation, where a unbuffered channel is still blocking at the end of the program.
The partner is a corresponding send or receive operation, which communicated with another operation, but could communicated with the stuck operation instead, resolving the deadlock.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestSelectMultiple
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/SelectBlock/select_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/SelectBlock/select_test.go:37
```go
26 ...
27 
28 	}
29 }
30 
31 // TestSelectMultiple: multiple cases
32 func TestSelectMultiple(t *testing.T) {
33 	ch1 := make(chan int)
34 	ch2 := make(chan int)
35 	go func() { ch1 <- 1 }()
36 	time.Sleep(5 * time.Millisecond)
37 	select {           // <-------
38 	case <-ch1:
39 	case <-ch2:
40 	case <-time.After(20 * time.Millisecond):
41 	}
42 }
43 
```


###  Channel: Send
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
The analyzer found a leak in the recorded trace.
The analyzer found a way to resolve the leak, meaning the leak should not reappear in the rewritten trace.

**Replaying was successful**.

It exited with the following code: 20

The replay was able to get the leaking unbuffered channel or select unstuck.

