# Fuzzing

One drawback of the implemented dynamic analysis is, that we can only
detect bugs, where all parts are in the executed code. To reduce this
problem, we use fuzzing. This involves running and analyzing a program or
test multiple times.

The simplest form would be to just let the program
or test run multiple times (black-box fuzzing). This may result in all possible
paths and schedules being executed at some times, but to get to this point, it may need to
run the program a large number of times running the same paths or schedules multiple times.

Another way would be to systematically explore all possible execution paths
and schedules. This will at some point find all possible bugs (assuming we
don't have changing inputs), but needs to run a prohibitively large number
of paths ([path explosion](https://en.wikipedia.org/wiki/Path_explosion)) and schedules.

We will look at a middle way, where we trie to influence the execution of
a program in such a way, that it increases the probability of exploring
new program paths, without forcing every single step.

Since the HB analysis implicitly checks all schedules for the executed paths,
we can focus on increasing the path cover and can do not need to explore all possible schedules.

The base template for fuzzing looks as follows:

```c
S := {ε}  // queue of schedules
while S != Ø and !timeout {
  σ := popMutation(S)
  s := runRecord(σ)
  e := getEnergy(s)
  for i := 1 to e {
    s_mut := mutateSchedule(s)
    if s_mut.isNew() {
      S := append(S, s)
    }
  }
}
```

We start with an set of schedules, which only contains the empty schedule.
We now start the fuzzing loop and continue, until no more schedules are
available or until a predefined timeout has been reached. In each loop,
we pick a schedule $\sigma$ from $S$ and run it. The empty schedule $\varepsilon$ means,
that we just let the program run without influencing it. Otherwise, we
make sure that the program run follows the given schedule.
Recording this run results in a new schedule $s$. We now determine this runs
energy (sometimes called score). This energy $e$ determines how interesting
the schedule was and how many mutations should be created. If the schedule
was not interesting, we can also get an energy of zero, creating no new
mutations. We then create create $e$ new mutations of $s$ and, if the
same mutation hasn't been seen before, add them to the
queue.

We will lock at multiple fuzzing approaches for concurrent bugs in go.\
For a bit of a background and ideas, see [here](fuzzing/background.md).

First we will integrate two already existing fuzzing approaches into our framework (GFuzz and GoPie).\
Then we will improve them using the HB relationships.\
Additionally we implement our own fuzzing approaches (Flow).

- [GFuzz](fuzzing/GFuzz.md)
- [GoPie](fuzzing/GoPie.md)
- [Flow](fuzzing/Flow.md)

[Here](./../examples/fuzzing/README.md) you can find some examples illustrating the
different approaches, and a comparison between the original and our GoPie
implementation when applying them to the GoBench benchmark.
