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

We will mainly look at two methods trying to find a middle ground between the
two extremes. For a bit of a background and ideas, see [here](fuzzing/background.md).

Since the HB analysis implicitly checks all schedules for the executed paths,
we only need to increase the path cover and can neglect the schedule cover.

The first method we look at is the [select based fuzzing](fuzzing/selectBased.md).
This methods focuses at selects at the main path branching point in
concurrent go programs.

The second method is the [order based fuzzing](fuzzing/orderBased.md).
This method tries to directly reorder the schedule and by doing so, tries to
unlock further execution paths (this is not implemented yet).