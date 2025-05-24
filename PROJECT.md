# Trace-based capture and replay in Go

Understanding and reproducing the behavior of concurrent programs is
notoriously difficult due to the nondeterministic nature of thread or
goroutine scheduling.

This project presents an implementation of a tracing and replay mechanism for
Go, built by modifying the Go runtime itself, rather than instrumenting user code.

By capturing key scheduling and synchronization events during execution, our
system enables deterministic replay of concurrent Go programs.

The primary goal of this work is to provide a general-purpose foundation for
deterministic replay in Go, which can be used to support a wide range of
analyses and tooling.

One particularly compelling application is the detection and diagnosis of
concurrency bugs such as send on closed channel and deadlocks, which often
depend on rare and hard-to-reproduce execution interleavings.

Our replay system allows such interleavings to be captured and replayed reliably,
facilitating debugging and testing workflows.

By integrating tracing and replay capabilities directly into the runtime, we
aim to maintain compatibility with unmodified Go programs and offer a
transparent, low-overhead mechanism for capturing execution behavior.
This approach lays the groundwork for tools that rely on deterministic
execution, including systematic testing and dynamic analysis.

## Achievements

For the project, we implemented the following parts

- Tracing of concurrency primitives in Go programs
  - routine local tracing
  - minimal need of program instrumentation by use of modified go runtime
- Replay of concurrent Go programs
  - Ability to execute program scheduling based on recorded and/or modified traces
  - minimal need of program instrumentation by use of modified go runtime
- Implementation of [GFuzz](./doc_proj/relatedWorks/PaperAndTools/Fuzzing/GFuzz.md) and [GoPie](./doc_proj/relatedWorks/PaperAndTools/Fuzzing/GoPie.md) fuzzing mechanisms and improvements for GoPie
  - GFuzz: integration of GFuzz idea into our framework
  - GoPie: integration of GoPie idea into our framework
  - GoPie+: improvements on GoPie idea
    - full coverage of all concurrency primitives, like mutex, channel, wait group, once,... (original GoPie only uses channel and mutex)
    - replay mechanism to guaranty that the program reaches the modified code block
    - consider scheduling chains resulting from mutated program runs
    - mutate partially executed operations
    - exclude superfluous mutations

## Implementations

The provided implementations consists of two parts.

The first is a modified Go runtime, provided in [go-patch](./go-patch/).
By directly modifying the runtime, including the implementations of the
concurrency operations, we provide recording and replay of concurrent
go programs without needing to change the recorded or replayed programs
(only adding a small header is required).

The second part is the [advocate](./advocate/) program. This program
starts, performs and supervises the recording, modifying and replaying
of given programs.

Replaying recorded program runs allows us to closely examine the recorded
execution and it allows us, by strategically modifying the trace, to
guide the execution into other possible execution schedules. This can be
used to confirm assumptions about possible executions, e.g. if an
analysis assumes that a certain schedule would cause the program to
panic, we can create and execute this schedule, proving the the
panic is actually possible. It also allows us to modify the execution
to discover new program code, not executed in the recorded run, making
(guided) fuzzing approaches possible.

Please note, that the current code base also includes code for trace analysis
using happens-before information and improvements on fuzzing approaches using
happens-before information. They have partially been build before or during the
project, but should not be seen as part of the project. The project only
focuses on the recoding and replay aspect, and using the replay with some simple
fuzzing approaches.

## Documentation

The documentation can be found in the [doc_proj](./doc_proj/) and [doc](./doc/) directories.
[doc_proj](./doc_proj/) contains all documentation that is relevant for the
project. [doc](./doc/) also includes the documentation for all other parts
that have been implemented into the Advocate framework.


The relevant section for this project are mostly the following:

- [Usage](./doc_proj/usage.md)
- [Toolchain](./doc_proj/toolchain.md)
- [Runtime](./doc_proj/runtime.md)
- [Execution, Recording and Trace](./doc_proj/recording.md)
- [Replay](./doc_proj/replay.md)
- [Fuzzing](./doc_proj/fuzzing.md)
- [Memory](./doc_proj/memory.md)
- [Related works](./doc_proj/relatedWorks.md)

But of course, feel free to take a look at the other things if you're interested.