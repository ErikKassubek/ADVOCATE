# Bug: Possible Send on Closed Channel

The analyzer detected a possible send on a closed channel.
Although the send on a closed channel did not occur during the recording, it is possible that it will occur, based on the happens before relation.
Such a send on a closed channel leads to a panic.

## Minimal Example
The following code is a minimal example to visualize the bug type. It is not the code where the bug was found.

```go
func main() {
    c := make(chan int)

    go func() {
        c <- 1          // <-------
    }()

    go func() {
        <- c
    }()

    close(c)            // <-------
}
```

## Test/Program
The bug was found in the following test/program:

- Test/Prog:  Test07
- File:  /home/erik/Uni/HiWi/ADVOCATE/examples/canonicalTests/tests/TN_SendAndRecOnClosed_test.go

## Bug Elements
The full trace of the recording can be found in the `advocateTrace` folder.

The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/erik/Uni/HiWi/ADVOCATE/examples/canonicalTests/tests/TN_SendAndRecOnClosed_test.go:22
```go
11 ...
12 
13 func n07() {
14 	c := make(chan int)
15 
16 	go func() {
17 		c <- 1
18 	}()
19 
20 	<-c
21 
22 	close(c)           // <-------
23 }
24 
```


###  Channel: Close
-> /home/erik/Uni/HiWi/ADVOCATE/examples/canonicalTests/tests/TN_SendAndRecOnClosed_test.go:27


## Replay
The bug is a potential bug.
The analyzer has tries to rewrite the trace in such a way, that the bug will be triggered when replaying the trace.

The rewritten trace can be found in the `rewritten_trace` folder.

**Replaying failed**.

It exited with the following code: 12

The replay got stuck during the execution.
No trace element was executed for a long tim.
This can be caused by a stuck replay.
Possible causes are:
    - The program was altered between recording and replay
    - The program execution path is not deterministic, e.g. its execution path is determined by a random number
    - The program execution path depends on the order of not tracked operations
    - The program execution depends on outside input, that was not exactly reproduced
