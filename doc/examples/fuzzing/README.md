# Fuzzing Examples

This directory contains examples to illustrate the advantages of the
different fuzzing approaches/modes and improvements.


## GoPie

GoPie directly mutates different operations. For GoPie only the order of
operations of channels and mutexes can be mutated. Examples where this is helpful
can be found for [mutexe](./goPie/mutex.go) and [channel](./goPie/channel.go).

When using the improved version GoPie+, we are also able to mutate the
order of other operations. An example where the order of atomic operations
may lead lead to a buggy code function being executed can be found
[here](./goPie/atomic.go).

GoPie+ also introduces a better replay system, forcing the exact replay
of the mutated run, until the muted position in the execution is reached.
This can reduce the possibility of the program getting lost or changing
before the mutated operations is reached. An example where this can be helpful can
be found [here](./goPie/replay.go).

Improving GoPie again to also use the HB relation (GoPieHB) has two main
benefits. For one, a bug does not need to be triggered directly, since
GoPieHB can infer possible bugs from the HB information. In GoPie and GoPie+,
this is not possible, which can lead to bugs being missed, even if the relevant
code is executed.

The second advantage is, that we can already filter out
impossible runs. Since especially GoPie has only a very limited view of the
program, it can create mutations that are impossible to execute, leading in
an unnecessary increase of the number of required runs. With the information, runs that,
based on the HB information, are impossible can directly be filtered out.
An example for this can be seen [here](./goPie/impossibleOrder.go).

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

In [GFuzz/hb1.go](./GFuzz/hb1.go) and [GFuzz/hb1.go](./GFuzz/hb1.go) we show
examples for how our improvements using the HB analysis may improve the
accuracy and speed of GFuzz.

## Flow

Flow changes the order of the two concurrent instances of the same
operation type on the same elements, where the first one succeeded, but the other did not.
We show three examples of this in

- [Flow/once.go](./Flow/once.go)
- [Flow/mutex.go](./Flow/mutex.go)
- [Flow/channel.go](./Flow/channel.go)

