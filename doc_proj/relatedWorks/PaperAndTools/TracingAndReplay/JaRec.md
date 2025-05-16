# JaRec: a portable record/replay environment for multi-threaded Java applications

[Georges, M. Christiaens, M. Ronsse, and K. De Bosschere, “Jarec: A portable record/re-
play environment for multi-threaded java applications,” Software: Practice and Experience,
vol. 34, no. 6, pp. 523–547, 2004. Available: https://onlinelibrary.wiley.com/doi/abs/10.1002/spe.579.](https://onlinelibrary.wiley.com/doi/abs/10.1002/spe.579)

## Summary
JaRec, is a record/replay system for Java. It replays multi-threaded, data-race free Java applications, by recording the order of synchronization operations, and by executing them in the same order during replay.

JaRec operates by capturing key events during the execution of a multi-threaded Java program. The recording process is implemented through a small recording agent that is attached to the JVM. This agent captures critical data related to thread execution and synchronization without modifying the program's code. This includes thread scheduling,synchronization events and shared memory accesses.

Once the execution is recorded, JaRec enables the deterministic replay of the program. The replay phase works by using the recorded log to re-execute the program with the exact same thread interleavings as the original run.

The replay engine ensures that the execution of threads during replay
follows the exact same order as recorded, regardless of the underlying
operating system's thread scheduler. JaRec controls the scheduling of
threads to ensure that the thread interleavings observed during the
original execution are reproduced. This deterministic scheduling is
achieved by forcing the threads to be scheduled according to the
sequence captured in the log file, overriding any randomness introduced
by the operating system’s scheduler.

The replay engine also ensures that the synchronization events (such as
lock acquisitions or releases) are executed in the same order as during
the original run. This guarantees that any deadlock or race condition
that occurred in the original execution can be reproduced accurately.
The replay engine also ensures that all memory accesses occur in the
same order as in the original run, allowing concurrency bugs related to
shared memory access to be faithfully reproduced.
