# Analysis

We have the ability, based on the program execution and the recorded trace to determine if a leak (goroutine that is still blocked when the program terminated) or panic took place in an execution.

For more information see here:

## Leak

### Goroutine leak

A goroutine leak is a goroutine, that is blocked when the program returns.

In our approach, we can identify potential leaks by checking for routines
where the last recorded event is a "request" event (tPost = 0).

### Non blocking Goroutine leak

In some cases it can happen, that a routine is still running at the end,
without it being blocked by one of the recorded operations. This can be
a desired behavior, but can also be a sign for undesired behavior.
For this reason, such cases are also detected.

To do this, we add an additional trace element into the trace, whenever
a routine terminates. In the analysis, we then traverse all routines and
check if there last element is such an termination element. If this is the
case, we have detected such a case. We then check if the penultimate element
(if it exists), has tPost = 0. In this case the situation is a go routine
leak as described above and is not again reported, to prevent double reports.
Otherwise, it is reported.

## Panics

We detect occurring panics directly in the runtime. For this, we add a function
into the implementation of the [panic](../../goPatch/src/runtime/panic.go#L744) function.

This panic is then always automatically called if a program panics.

Based on the type and content of the message in the panic, we can sort the
panics based on our interests into

- "send on closed channel"
- "close of closed channel"
- "close of nil channel"
- "sync: negative WaitGroup counter"
- "test timed out"
- "sync: RUnlock of unlocked RWMutex"
- "sync: Unlock of unlocked RWMutex"
- "sync: unlock of unlocked mutex"
- other

This info can then used during the replay to determine if an expected bug
has been triggered and for the recording and fuzzing to directly
find actual bugs.
