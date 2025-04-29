# Leak

## Goroutine leak

A goroutine leak is an indefinitely blocked goroutine.

In our approach, we can identify potential leaks by checking for goroutines
where the last recorded event is a "pre" event (tpost = 0).

For channel operation, we try to find a possible partner, which is
used in the trace reorder to get the operation unstuck.

For the other operation, no additional analysis besides finding them is done.
This is done in the following way. For each channel and each routine,
the last processed send and receive is recorded (the elements are processed
in the order of there execution in the trace).
If a stuck channel or select element is processed, we check if one of the elements
in the last processes send or receives is a possible partner. If a possible
partner is found, the stuck element with its partner is added to the
analysis results. If no partner can be found, the operation is added to
a list of stuck channel elements without partner $s$.
For each non stuck channel element and each case in a non stuck
select element, including not selected cases in those elements, we check
if the element would be a potential partner for one of the stuck elements.
If it is, the element and its partner are
added to the analysis result and the element is removed from $s$.
If all elements are processed, we traverse through $s$ and add a result
for an stuck channel element without possible partner to the analysis result
for each element in $s$.

For a leaking mutex $m$, we add the mutex and the last successful lock of this
mutex before $m$ to the analysis result.

For all other stuck elements, we only add the stuck element to the analysis
result.


## Analysis scenario: Non blocking Goroutine leak

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

There is no rewrite/replay for those cases.