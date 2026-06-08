# Leak: L09 - Leak on Sync.WaitGroup

The analyzer detected a leak on a sync.WaitGroup.
A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.WaitGroup wait is blocking, because the counter is not zero.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach7504
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/7504/cockroach7504_test.go
- Trace: advocateTrace_18

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Waitgroup: Wait
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/4/blocking/cockroach/6181/cockroach6181_test.go#62
```go
51 ...
52 
53 		var wg sync.WaitGroup
54 		for i := 0; i < 3; i++ {
55 			wg.Add(1)
56 			go func() { // G2,G3,...
57 				doLookupWithToken(db.cache)
58 				wg.Done()
59 			}()
60 		}
61 		wg.Wait()
62 	}           // <-------
63 	pauseLookupResumeAndAssert()
64 }
65 
66 /// G1 									G2							G3					...
67 /// testRangeCacheCoalescedRquests()
68 /// initTestDescriptorDB()
69 /// pauseLookupResumeAndAssert()
70 /// return
71 /// 									doLookupWithToken()
72 ///																 	doLookupWithToken()
73 
74 ...
```


## Replay
**Replaying was not run**.

