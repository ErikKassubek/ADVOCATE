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
- [New Channel](trace/newChannel.md)
- [Fork](trace/fork.md) (Start of new routine)
- [Return of Routine](trace/routineEnd.md)

Additionally, the replay can add in additional [markers](trace/replay.md).

## Trace Recording
To run the recording, we need to add the following header to the main function
or the test that is analyzed:

```go
// ======= Preamble Start =======
  advocate.InitTracing()
  defer advocate.FinishTracing()
// ======= Preamble End =======
```

When the toolchain is used, this is done automatically.

To routines locally record the trace, we add an new variable `advocateRoutineInfo` into the [g struct](../go-patch/src/runtime/runtime2.go#L517).
This struct is automatically created for each routine.
This variable (defined [here](../go-patch/src/runtime/advocate_routine.go#L28)), stores the routine id,
the maximum id of any element used in this routine and the
Trace as list of elements. They are set [when the routine is started](../go-patch/src/runtime/proc.go#L5080).

For each struct representing one of the recorded operations (except for fork and atomic operations), we add a field `id`, to store the `id` for this element. When a recorded function
is executed, we first check if this id was already set. If it was not, we set a new id. We want to minimize the number of
global counters we need to use. We therefor construct the new
id as $routine.id\cdot1000000000 + routine.maxObjectId$ and then increase the maxObjectId field of the routine. For atomics, we use the memory position of the value as id.

For each recorded operation a Pre and sometimes a Post function is implemented (multiple operations on the same type may share a Post function). The Pre function is called when the operation
is started, but before it executes. The Post function is called after it finished executing.

The Pre function record that the operation was executed, as well as all required information that are already available (for the lest if required information, check out the type specific information linked above). The Pre function
also always records the location where the operation was recorded. This is (except for fork, see [here](./trace/fork.md#implementation)) done using the [runtime.Caller](https://pkg.go.dev/runtime#Caller) function.

The Post function then records the successful completion of the
operation (by setting tPost to not zero) as well as all information that was not available at the beginning, e.g. if a trylock was successful or not or which select case was executed.

When the program execution has finished, it will create a folder `advocateTrace`
in which it stores the trace files. For each routine, one trace file will be
generated. In it, each line contains the information about one recorded
event. The events are sorted by the time when the operations was executed.

### Timer
To reconstruct the global trace from the recorded local traces, we need a consistent
timer.

The timer is implemented using an atomic variable, which is incremented every
time a pre or post event is created. Unfortunately this needed to be implemented
as a global counter, to get a consistent time value. Experiments with
more local timer values have been made. The two methods we tried consisted in using the runtime.cputicks() and the runtime.nanotime() functions.

The cputicks function is described by the go trace team as
"On most platforms, this function queries the CPU for a tick count with a single instruction. (Intuitively a "tick" goes by roughly every CPU clock period, but in practice this clock usually has a constant rate that's independent of CPU frequency entirely.) [...] Unfortunately, many modern CPUs don't provide such a clock that is stable across CPU cores, meaning even though cores might synchronize with one another, the clock read-out on each CPU is not guaranteed to be ordered in the same direction as that synchronization. This led to traces with inconsistent timestamps." [^1]

Additionally, the signature of the implementation notes:
"careful: cputicks is not guaranteed to be monotonic! In particular, we have noticed drift between cpus on certain os/arch combinations. See issue 8976." [^2]

For this reason, we cannot guarantee that a trace with this form of timestamp reflects the actual execution order.

The nanotime() returns a time value from the operating system.

[^1]: M. Knyszek. "Execution tracer overhaul". https://github.com/golang/proposal/blob/master/design/60773-execution-tracer-overhaul.md (Accessed 2025-03-29)
[^2]: [runtime/cputicks.go](../go-patch/src/runtime/cputicks.go#L11)