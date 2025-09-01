# Go Tracer

- Links:
  - [Documentation](https://pkg.go.dev/runtime/trace)
  - [More powerful Go execution traces](https://go.dev/blog/execution-traces-2024)

The Go tracer is part of the go runtime. It provides detailed insights into the
execution flow of Go Programs.

The tracer records the following operations [^1]

```
traceEvNone traceEv = iota // unused

// Structural events.
traceEvEventBatch // start of per-M batch of events [generation, M ID, timestamp, batch length]
traceEvStacks     // start of a section of the stack table [...traceEvStack]
traceEvStack      // stack table entry [ID, ...{PC, func string ID, file string ID, line #}]
traceEvStrings    // start of a section of the string dictionary [...traceEvString]
traceEvString     // string dictionary entry [ID, length, string]
traceEvCPUSamples // start of a section of CPU samples [...traceEvCPUSample]
traceEvCPUSample  // CPU profiling sample [timestamp, M ID, P ID, goroutine ID, stack ID]
traceEvFrequency  // timestamp units per sec [freq]

// Procs.
traceEvProcsChange // current value of GOMAXPROCS [timestamp, GOMAXPROCS, stack ID]
traceEvProcStart   // start of P [timestamp, P ID, P seq]
traceEvProcStop    // stop of P [timestamp]
traceEvProcSteal   // P was stolen [timestamp, P ID, P seq, M ID]
traceEvProcStatus  // P status at the start of a generation [timestamp, P ID, status]

// Goroutines.
traceEvGoCreate            // goroutine creation [timestamp, new goroutine ID, new stack ID, stack ID]
traceEvGoCreateSyscall     // goroutine appears in syscall (cgo callback) [timestamp, new goroutine ID]
traceEvGoStart             // goroutine starts running [timestamp, goroutine ID, goroutine seq]
traceEvGoDestroy           // goroutine ends [timestamp]
traceEvGoDestroySyscall    // goroutine ends in syscall (cgo callback) [timestamp]
traceEvGoStop              // goroutine yields its time, but is runnable [timestamp, reason, stack ID]
traceEvGoBlock             // goroutine blocks [timestamp, reason, stack ID]
traceEvGoUnblock           // goroutine is unblocked [timestamp, goroutine ID, goroutine seq, stack ID]
traceEvGoSyscallBegin      // syscall enter [timestamp, P seq, stack ID]
traceEvGoSyscallEnd        // syscall exit [timestamp]
traceEvGoSyscallEndBlocked // syscall exit and it blocked at some point [timestamp]
traceEvGoStatus            // goroutine status at the start of a generation [timestamp, goroutine ID, M ID, status]

// STW.
traceEvSTWBegin // STW start [timestamp, kind]
traceEvSTWEnd   // STW done [timestamp]

// GC events.
traceEvGCActive           // GC active [timestamp, seq]
traceEvGCBegin            // GC start [timestamp, seq, stack ID]
traceEvGCEnd              // GC done [timestamp, seq]
traceEvGCSweepActive      // GC sweep active [timestamp, P ID]
traceEvGCSweepBegin       // GC sweep start [timestamp, stack ID]
traceEvGCSweepEnd         // GC sweep done [timestamp, swept bytes, reclaimed bytes]
traceEvGCMarkAssistActive // GC mark assist active [timestamp, goroutine ID]
traceEvGCMarkAssistBegin  // GC mark assist start [timestamp, stack ID]
traceEvGCMarkAssistEnd    // GC mark assist done [timestamp]
traceEvHeapAlloc          // gcController.heapLive change [timestamp, heap alloc in bytes]
traceEvHeapGoal           // gcController.heapGoal() change [timestamp, heap goal in bytes]

// Annotations.
traceEvGoLabel         // apply string label to current running goroutine [timestamp, label string ID]
traceEvUserTaskBegin   // trace.NewTask [timestamp, internal task ID, internal parent task ID, name string ID, stack ID]
traceEvUserTaskEnd     // end of a task [timestamp, internal task ID, stack ID]
traceEvUserRegionBegin // trace.{Start,With}Region [timestamp, internal task ID, name string ID, stack ID]
traceEvUserRegionEnd   // trace.{End,With}Region [timestamp, internal task ID, name string ID, stack ID]
traceEvUserLog         // trace.Log [timestamp, internal task ID, key string ID, stack, value string ID]

// Coroutines.
traceEvGoSwitch        // goroutine switch (coroswitch) [timestamp, goroutine ID, goroutine seq]
traceEvGoSwitchDestroy // goroutine switch and destroy [timestamp, goroutine ID, goroutine seq]
traceEvGoCreateBlocked // goroutine creation (starts blocked) [timestamp, new goroutine ID, new stack ID, stack ID]

// GoStatus with stack.
traceEvGoStatusStack // goroutine status at the start of a generation, with a stack [timestamp, goroutine ID, M ID, status, stack ID]

// Batch event for an experimental batch with a custom format.
traceEvExperimentalBatch // start of extra data [experiment ID, generation, M ID, timestamp, batch length, batch data...]
```

This means, it gives the user insight into

- Goroutine Events
  - Create
  - Schedule
  - Blocked
  - Unblocked
  - Finished
- Garbage Collection
- Thread Events
- Heap and Memory Events
- System Calls
- ...

Those events give a good insights to analyze the project performance
or detect bottlenecks, but for our use, not enough events are detected.
We can infer channels communication or mutex operations that
cause the routines to block, but if a communication or mutex operations
is directly executed without blocking the routine first, we cannot get the
information from the trace. The same is true for all other operations
that do not block the execution of a go routine.

It would have been able to modify the tracer to also record the
operations we are interested in. Similar to our tracer, the go tracer
contains functions `eventWriter` and `event` [^2], which are called to
record one trace event. We could have amended this function and the list of
trace events. By then adding the `eventWrite().event()` functions into the different
operations we want to record, it would have been possible to record the
different operations with the trace.

[^1] [runtime/traceEvents.go](../../goPatch/src/runtime/traceevent.go)\
[^2] [runtime/traceEvents.go:121](../../goPatch/src/runtime/traceevent.go#L121)
