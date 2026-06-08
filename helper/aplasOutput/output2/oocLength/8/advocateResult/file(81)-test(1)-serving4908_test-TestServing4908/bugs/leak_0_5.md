# Leak: L09 - Leak on Sync.WaitGroup

The analyzer detected a leak on a sync.WaitGroup.
A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.
A Leak could potentially resolve itself, if the program would run longer.
A sync.WaitGroup wait is blocking, because the counter is not zero.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestServing4908
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/nonblocking/serving/4908/serving4908_test.go
- Trace: advocateTrace_82

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Waitgroup: Wait
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/8/nonblocking/serving/3068/serving3068_test.go#78


## Replay
**Replaying was not run**.

