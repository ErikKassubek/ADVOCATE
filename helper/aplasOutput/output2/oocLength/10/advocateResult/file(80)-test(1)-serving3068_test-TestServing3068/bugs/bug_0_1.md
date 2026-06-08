# Bug: A01 - Actual Send on Closed Channel

During the execution of the program, a send on a closed channel occurred.
The occurrence of a send on closed leads to a panic.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestServing3068
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/nonblocking/serving/3068/serving3068_test.go
- Trace: advocateTrace_81

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Send
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/nonblocking/serving/3068/serving3068_test.go#44
```go
33 ...
34 
35 			}
36 		}()
37 	}
38 
39 	return i
40 }
41 
42 func (i *impl) Go(w func()) {
43 	i.wg.Add(1)
44 	i.workCh <- w           // <-------
45 }
46 
47 func (i *impl) Wait() {
48 	i.once.Do(func() {
49 		close(i.workCh)
50 
51 		go func() {
52 			i.wg.Wait()
53 		}()
54 	})
55 
56 ...
```


## Replay
**Replaying was not run**.

