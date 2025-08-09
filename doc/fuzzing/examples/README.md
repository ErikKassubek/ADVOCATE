# Fuzzing Examples

This directory contains examples to illustrate the advantages of the
different fuzzing approaches/modes and improvements.


## GoPie

GoPie directly mutates different operations. For GoPie only the order of
operations of channels and mutexes can be mutated. Examples where this is helpful
can be found for [mutexe](./goPie/mutex.go) and [channel](./goPie/channel.go).

When using the improved version GoCR, we are also able to mutate the
order of other operations. An example where the order of atomic operations
may lead lead to a buggy code function being executed can be found
[here](./goPie/atomic.go).

GoCR also introduces a better replay system, forcing the exact replay
of the mutated run, until the muted position in the execution is reached.
This can reduce the possibility of the program getting lost or changing
before the mutated operations is reached. An example where this can be helpful can
be found [here](./goPie/replay.go).

To show, that our implementation of GoPie can replicate the results of the
original implementation, we have applied it to the GoBench benchmark,
and compared the results to the numbers given in the GoPie paper.
A summary of this can be found [here](./goPie/GoBench.md).

## GFuzz

GFuzz forces the execution of select cases, which are possible but may be
unlikely to be directly executed during an unguided run.

In [GFuzz/select.go](./GFuzz/select.go), a typical example of a select with two
cases is presented, where one is much less likely to be executed, but not impossible.
If this case leads to a bug, an unguided dynamic analysis will most likely
miss it.

In [GFuzz/selectWithDefault.go](./GFuzz/selectWithDefault.go) a similar
example is shown, in which the triggering of the default is the most likely
scenario, but triggering another case could lead to a bug.

In [GFuzz/select2.go](./GFuzz/select2.go), a program where only a specific
combination of chosen cases in multiple selects leads to a bug.