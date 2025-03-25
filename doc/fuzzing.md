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
of paths ([path explosion](https://en.wikipedia.org/wiki/Path_explosion)).

We will lock at multiple fuzzing approaches for concurrent bugs in go.\
For a bit of a background and ideas, see [here](fuzzing/background.md).


Since the HB analysis implicitly checks all schedules for the executed paths,
we only need to increase the path cover and can neglect the schedule cover.

First we will integrate two already existing fuzzing approaches into our framework.\
Later we will improve them using the HB relationships (not yet implemented).\
Additionally we will implement our own fuzzing approaches.