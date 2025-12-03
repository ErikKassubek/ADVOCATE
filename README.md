# AdvocateGo

## What is AdvocateGo

AdvocateGo is an analysis tool for concurrent Go programs.
It tries to detects concurrency bugs and gives diagnostic insight.

AdvocateGo tries to detect the following situations:

- A01: "Actual Send on Closed Channel",
- A02: "Actual Receive on Closed Channel",
- A03: "Actual Close on Closed Channel",
- A04: "Actual close on nil channel",
- A05: "Actual negative Wait Group",
- A06: "Actual unlock of not locked mutex",
- A07: "Partial Deadlock",
- A08: "Concurrent Receive",
- A09: "Select Case without Partner",
- P01: "Possible Send on Closed Channel",
- P02: "Possible Receive on Closed Channel",
- P03: "Possible Negative WaitGroup cCounter",
- P04: "Possible unlock of not locked mutex",
- P05: "Possible cyclic deadlock",
- L00: "Leak",
- L01: "Leak on unbuffered channel with possible partner",
- L02: "Leak on unbuffered channel without possible partner",
- L03: "Leak on buffered Channel with possible partner",
- L04: "Leak on buffered Channel without possible partner",
- L05: "Leak on nil channel",
- L06: "Leak on select with possible partner",
- L07: "Leak on select without possible partner",
- L08: "Leak on sync.Mutex",
- L09: "Leak on sync.WaitGroup",
- L10: "Leak on sync.Cond",
- L11: "Leak on channel or select on context",

Additionally it is able to record and deterministically replay
executions of concurrent GO programs.

## Modes

Advocate provides 4 different modes:

- record: record the execution of a program or test into a trace
- replay: given a trace file, execute a program in such a way, that it follows the trace
- analysis: record a program and analyze the recorded trace to detect potential concurrency bugs. If a potential bug is found, rewrite the trace in such a way that the bug is triggered and replay this trace to confirm that the bug is possible.
- fuzzing: Apply different fuzzing approaches to increase the reach of the analysis.

## Usage

> [!WARNING]
> This program currently only runs / is tested under Linux

> [!IMPORTANT]
> advocate is implemented for go version 1.25.
> Make sure, that the program does not choose another version/toolchain and is compatible with go 1.25.
> The output `package advocate is not in std ` or similar indicates a problem with the used version.

For an explanation on how to use advocate, see [here](./doc/usage.md).

## Documentation

A detailed description of how advocate works can be found in the [doc](doc) folder.