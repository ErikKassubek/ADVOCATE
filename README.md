# AdvocateGo

## What is AdvocateGo

AdvocateGo is an analysis tool for concurrent Go programs.
It detects concurrency bugs and gives diagnostic insight.

Furthermore it is also able to produce traces which can be fed back into the program in order to experience the predicted bug.

AdvocateGo tries to detect the following situations:

- A00: Unknown panic
- A01: Send on closed channel
- A02: Receive on closed channel (only warning)
- A03: Close on closed channel
- A04: Close on nil channel
- A05: Negative wait group
- A06: Unlock of not locked mutex
- A07: Concurrent recv
- P01: Possible send on closed channel
- P02: Possible receive on closed channel (only warning)
- P03: Possible negative waitgroup counter
- L01: Leak on unbuffered channel with possible partner
- L02: Leak on unbuffered channel without possible partner
- L03: Leak on buffered channel with possible partner
- L04: Leak on buffered channel without possible partner
- L05: Leak on nil channel
- L06: Leak on select with possible partner
- L07: Leak on select without possible partner (includes nil channels)
- L08: Leak on mutex
- L09: Leak on waitgroup
- L10: Leak on cond

## Documentation

A detailed description of the inner workings can be found in the [doc](doc) folder (currently in the process of being rewritten and therefore not complete).

## Usage

> [!WARNING]
> This program currently only runs / is tested under Linux

> [!IMPORTANT]
> ADVOCATE is implemented for go version 1.24.
> Make sure, that the program does not choose another version/toolchain and is compatible with go 1.24.
> The output `package advocate is not in std ` or similar indicates a problem with the used version.

For an explanation on how to use ADVOCATE, see [here](./doc/usage.md).

## Warning

It is the users responsibility of the user to make sure, that the input to
the program, including e.g. API calls are equal for the recording and the
tracing. Otherwise the replay is likely to get stuck.

Do not change the program code between trace recording and replay. The identification of the operations is based on the file names and lines, where the operations occur. If they get changed, the program will most likely block without terminating. If you need to change the program, you must either rerun the trace recording or change the effected trace elements in the recorded trace.
This also includes the adding of the replay header. Make sure, that it is already in the program (but commented out), when you run the recording.
