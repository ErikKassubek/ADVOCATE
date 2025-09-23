# Partial deadlock in an execution

While a total deadlock, involving all routines, can directly be detected
by the runtime, this is not the case for a partial deadlock, where only
a subset of the routines is blocked. For this reason, we implement our
own partial deadlock detection.

## Idea

The partial deadlock detection is based on [^1]. We utilize the
garbage collector to detect waiting elements, with no reference in any
other alive routines, that could unblock the waiting routines. For
a more depth explanation, see [here](deadlockInExecution.pdf).

## Usage

The detection is started by adding the following line at the beginning
of the program code:

```go
advocate.DetectPartialDeadlock(1000)
```

For this to work, the program must be run with the [modified go runtime](../../goPatch/).

This will start a background routine, that checks repeatedly for partial deadlocks.
The parameter gives the time in ms between two checks (in most cases a check
every second or so is enough, assuming the program runs for more than a second).
To perform only one check, set the parameter to 0,

If, for some reason, you want to stop a started looping detection, you can do
this with

```go
StopPartialDeadlockDetection()
```

This will finish the currently running check end then end a loop. The loop
can be started and stopped multiple time. Calling stop while the
partial deadlock detection is running has no effect.

When using advocate in the fuzzing mode, the partial deadlock detection
is automatically started.

## Output

If the detector finds a partial deadlock, it will print a message to
the terminal in the form

```
DEADLOCK@[routID]@[pos]@[waitReason]
```

with

- routID: the id of the blocked routine. If run while advocate recording is run
(e.g. in tracing or fuzzing mode), this is equal to the id of the recorded trace
- pos: position of the blocking element in the code in the form [file]:[line]
- blocking operation type and operation in the form [type]:[operation], e.g. mutex:lock,
chan:send





[^1]: G.-V. Saioc, I.-T. A. Lee, A. Møller, and M. Chabbi, “Dynamic Partial
Deadlock Detection and Recovery via Garbage Collection,” in Proceedings
of the 30th ACM International Conference on Architectural Support for
Programming Languages and Operating Systems, Volume 2, New York, NY,
USA: Association for Computing Machinery, Mar. 2025, pp. 244–259, isbn:
979-8-4007-1079-7.