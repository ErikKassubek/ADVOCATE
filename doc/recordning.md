# Execution, Recording and Trace

ADVOCATE uses dynamic analysis, meaning it runs the code that should be analyzed,
records the relevant information about this execution and tries to deduce
potential concurrency problems from this trace. For information about the
analysis, see [here](analysis.md).

The recording of the trace is implemented in the [modified runtime](../go-patch/src/runtime).
It can be found in the [advocate_trace.go](../go-patch/src/runtime/advocate_trace.go) file and the
advocate_trace_....go files in the same directory.

To run the recording, a special header is added to the running program by
the toolchain.

For each of the operations we want to record, additional function calls
have been added to the operations implementations.
We record the following types and operations:

- [Channel](trace/channel.md): Send, Receive, Close
- [Select](trace/select.md)
- [Mutex](trace/mutex.md): Lock, RLock, TryLock, TryRLock, Unlock, RUnlock
- [WaitGroup](trace/waitGroup.md): Add, Done
- [Once](trace/once.md): Do
- [Conditional Variable](trace/conditionalVariables.md): Wait, Signal, Broadcast
- [Atomics](trace/atomics.md): Load, Store, Add Swap, CompareAndSwap
- [Creation of a recorded type](trace/new.md)
- [Fork](trace/fork.md) (Start of new routine)
- [Return of Routine](trace/routineEnd.md)

Additionally, the replay can add in additional [markers](trace/replay.md).

When the program execution has finished, it will create a folder `advocateTrace`
in which it stores the trace files. For each routine, one trace file will be
generated. In it, each line contains the information about one recorded
event. The events are sorted by the time when the operations was executed.