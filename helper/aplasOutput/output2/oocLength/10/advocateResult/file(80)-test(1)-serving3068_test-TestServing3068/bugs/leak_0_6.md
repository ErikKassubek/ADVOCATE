# Leak: L09 - Leak on Sync.WaitGroup

The analyzer detected a leak on a sync.WaitGroup.
A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.WaitGroup wait is blocking, because the counter is not zero.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestServing3068
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/nonblocking/serving/3068/serving3068_test.go
- Trace: advocateTrace_81

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Waitgroup: Wait
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/nonblocking/serving/3068/serving3068_test.go#73
```go
62 ...
63 
64 		for i := 0; i < n; i++ {
65 			p.Go(func() {
66 				atomic.AddInt32(&cntExecuted, 1)
67 			})
68 			time.Sleep(10 * time.Millisecond)
69 			wg.Done()
70 		}
71 	}()
72 	p.Wait()
73 	wg.Wait()           // <-------
74 	if cntExecuted == n {
75 		t.Error("Not all items were expected to execute")
76 	}
77 }
78 
```


## Replay
**Replaying was not run**.

