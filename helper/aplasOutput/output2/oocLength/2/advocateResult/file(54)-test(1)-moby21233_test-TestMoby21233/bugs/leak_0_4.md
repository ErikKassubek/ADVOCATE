# Leak: L07 - Block on select

The analyzer detected a Block on a select.
A Block on a select is a situation, where a select is still blocking at the end of the program.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMoby21233
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/2/blocking/moby/21233/moby21233_test.go
- Trace: advocateTrace_55

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Select:
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/2/blocking/moby/21233/moby21233_test.go#89
```go
78 ...
79 
80 		for {
81 			t.mu.Lock()
82 			t.mu.Unlock()
83 			if rand.Int31n(2) >= 1 {
84 				progressOutput.WriteProgress(lastProgress)
85 			}
86 			if done {
87 				return
88 			}
89 			select {           // <-------
90 			case <-w.signalChan:
91 			case <-w.releaseChan:
92 				done = true
93 				select {
94 				default:
95 				}
96 			}
97 		}
98 	}()
99 	return w
100 
101 ...
```


## Replay
**Replaying was not run**.

