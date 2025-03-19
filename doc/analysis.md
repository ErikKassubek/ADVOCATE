# Analysis

For the analysis, we perform a happens bevor analysis using vector clocks.
For a detailed description of the happens before relation,
and the calculation of the vector clocks, see [here](analysis/hb.md).

Based on this relationship, we then try to find possible bugs or warnings.

We try to find

- [send/recv on closed channel](analysis/comOnClosed.md)
- [negative wait group counter](analysis/doneBeforeAdd.md)
- [unlock of not locked mutex](analysis/unlockBeforeLock.md)
- [cyclick deadlocks](analysis/cyclicDeadlock.md)
- [leaks](analysis/leak.md)
- [concurrent recv on the same channel](analysis/concurrentReceive.md)

To get an overview about possible bugs and how they are represented
in the results, see [here](analysis/results.md)