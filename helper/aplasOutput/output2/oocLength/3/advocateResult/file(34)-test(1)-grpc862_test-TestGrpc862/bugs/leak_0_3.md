# Leak: L00 - Leak

The analyzer detected a leak.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
A Leak could potentially resolve itself, if the program would run longer.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestGrpc862
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/3/blocking/grpc/862/grpc862_test.go
- Trace: advocateTrace_35

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  
## Replay
**Replaying was not run**.

