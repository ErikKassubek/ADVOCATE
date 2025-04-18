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

We wanted to implement the timer routine local if possible. For this, we looked at two possible time functions in the go runtime. Both of them have there problems

The `cputicks` function is described by the go tracer team as\
"On most platforms, this function queries the CPU for a tick count with a single instruction. (Intuitively a "tick" goes by roughly every CPU clock period, but in practice this clock usually has a constant rate that's independent of CPU frequency entirely.) [...] Unfortunately, many modern CPUs don't provide such a clock that is stable across CPU cores, meaning even though cores might synchronize with one another, the clock read-out on each CPU is not guaranteed to be ordered in the same direction as that synchronization. This led to traces with inconsistent timestamps." [^1]

Additionally, the signature of the implementation notes:\
"careful: cputicks is not guaranteed to be monotonic! In particular, we have noticed drift between cpus on certain os/arch combinations. See issue 8976." [^2]

For this reason, we cannot guarantee that a trace with this form of timestamp reflects the actual execution order.

The nanotime() returns a time value from the operating system. It it consistent over all routines.

Here the problem lyes in the precision of those counters. Atomic operations like Load and Store only take a few nano seconds to execute. Assume we have the following program snippet.

```go
var a atomic.Uint64
go func() {
  a.Store(1)
}()
go func() {
  x := a.Load()
}
```

If the following code in the second routine depends on the value of `x`, it is necessary
to determine the exact order of the Load and Store operations, to get an accurate analysis and especially replay.
If we use the nanosecond counter as timer and the precision of the timer is to small, this may result in
unclear orders.


For Linux based systems, this uses `CLOCK_MONOTONIC` (normally defined in `/usr/include/time.h`) with a precision of, in most cases, `1 ns` (can by checked by running `clock_getres(CLOCK_MONOTONIC, &res); printf("%ld\n", res.tv_nsec)` as a c program).
It is therefore enough. The problem, lies in the use in windows and macos. For windows, the `QueryPerformanceCounter` is used. According to [^3], this only has a resolution of about `100 ns`. This could lead to a situation, where the two atomic
operations receive the same timer value, resulting in an incorrect replay.

For this reason, we used the following approach:\
For Linux system we use the `nanotime` function as our timer values.\
For all other systems, we use a global atomic variable, that is increased every time a timer is requested (`return advocateGlobalCounter.Add(1)`).

Those two different functions are implemented by using two files, annotated with `//go:build linux` for the Linux function and `//go:build !linux` for all other systems.

[^1]: M. Knyszek. "Execution tracer overhaul". https://github.com/golang/proposal/blob/master/design/60773-execution-tracer-overhaul.md (Accessed 2025-03-29)\
[^2]: [runtime/cputicks.go](../go-patch/src/runtime/cputicks.go#L11)\
[^3]: S, White et al. "Acquiring high-resolution time stamps". https://learn.microsoft.com/en-us/windows/win32/sysinfo/acquiring-high-resolution-time-stamps#resolution-precision-accuracy-and-stability (Accessed 2025-03-29)