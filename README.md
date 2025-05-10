# AdvocateGo

## What is AdvocateGo

AdvocateGo is an analysis tool for concurrent Go programs.
It tries to detects concurrency bugs and gives diagnostic insight.

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
- L00: Leak on routine with unknown cause
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
> advocate is implemented for go version 1.24.
> Make sure, that the program does not choose another version/toolchain and is compatible with go 1.24.
> The output `package advocate is not in std ` or similar indicates a problem with the used version.

For an explanation on how to use advocate, see [here](./doc/usage.md).

## Documentation

A detailed description of how advocate works can be found in the [doc](doc) folder.