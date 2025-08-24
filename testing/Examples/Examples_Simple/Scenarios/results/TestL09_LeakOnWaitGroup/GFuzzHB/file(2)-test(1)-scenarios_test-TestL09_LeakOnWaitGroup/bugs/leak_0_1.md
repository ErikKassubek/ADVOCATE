# Leak: Leak on sync.WaitGroup

The analyzer detected a leak on a sync.WaitGroup.
A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.
A sync.WaitGroup wait is blocking, because the counter is not zero.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestL09_LeakOnWaitGroup
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Waitgroup: Wait
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Own/Scenarios/scenarios_test.go:166
```go
155 ...
156 
157 	mu.Lock() // L08: blocked
158 }
159 
160 func TestL09_LeakOnWaitGroup(t *testing.T) {
161 	var wg sync.WaitGroup
162 	wg.Add(1)
163 	go func() {
164 		time.Sleep(100 * time.Millisecond) // never calls Done()
165 	}()
166 	wg.Wait() // L09           // <-------
167 }
168 
169 func TestL10_LeakOnCond(t *testing.T) {
170 	mu := sync.Mutex{}
171 	cond := sync.NewCond(&mu)
172 	go func() {
173 		time.Sleep(50 * time.Millisecond)
174 		// no cond.Signal()
175 	}()
176 	mu.Lock()
177 
178 ...
```


## Replay
**Replaying was not run**.

