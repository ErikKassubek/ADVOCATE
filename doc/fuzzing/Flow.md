# Flow Changing Operations

Our bug detection is based on HB analysis. It is therefore not really useful
in the fuzzing, to create mutations, that lead to the same happens before
relations, since they do not produce new information.

Instead we want to focus on mutation techniques that explore new execution paths.
On of them, that is already implemented is the mutation of the selects,
as based on GoFuzz. Tho following primitive are of interest.

One of those would e.g. be `once`. If we have two `Do` on the same `once` in the
trace, were one of them was executed and the other was not, and both of them
are concurrent based on the HB analysis, one can create a mutation to force
this execution.

Similar scenarios can be created e.g. for `try-locks`.

Another situation we try to capture, would be for code of the form
```go
c := make(chan int, 2)

go func() {
  c <- 0
}()

go func() {
  c <- 1
}

if a := <-c; a == 0 {
  // code with bug
}
```
Here the execution of the code part containing the bug depends on the order
in which the two sends are executed. We could try to implement mutations,
where we change the order of channel operations of the same type and on the same
channel if the two operations are HB concurrent. Since this should not happen
that often, it should be possible to implement this without needing to create
massive amounts of mutations, and we would be able to increase the number of
paths we explore.

## Implementations

We implemented the flow mutation for `once`, `mutex`, `channel`.
While performing the analysis, we look for operations of on the same element
and of the same operation (e.g. Do on the same once, send on the same channel),
that happen concurrently, based on the HB information. We hereby focus mainly
on situations, where the first of those executions was successful (executed
Do functions, successful acquire in tryLock, executed send, ...), but the
second was not.\
If this is the case, we will create a run, where the order of those
are switched.

It would be possible to directly write a new trace, where the concurrent
once, mutex, send and receive are reversed. But here we use a simple
method to reduce the overhead. For each of the pairs of operations we
wish to reorder, we store the operations, that in the recorded trace was first
in a file.

When an operation tries to execute an operations, it will first check if
it is in this file. If this is the case, the execution of the operation
is paused for a predefined number of second. After this timer has passed it
will be released and the operation execution. This gives the other operation
time to execution first, therefore reversing the order if the operations.
This also implicitly contains a timeout. If the operation that should be
execute first has not been executed, e.g. because the oder of the operations
cannot be changed for some reason, the waiting operation will continue its
execution, therefore continuing the program without the possibility of it getting stuck.

## Comparison to GoPie

Lets look at the following example:

```go
c := make(chan int)
on := sync.Once{}

numberRuns := 1000

go func() {
  on.Do(codeWithoutPosBug)  // do 1
}()

// create some noise for GoPie
go func() {
  for i := range numberRuns {
      c <- 1
  }
}()
for i := range numberRuns {
    <-c
}

on.Do(codeWithPosBug)  // do 2
```
Because of the channels, `do 1` will normally be executed before the
`do 2` and the code with the possible bug will therefore not be executed. But from the HB relation we know, that they are concurrent and the `codeWithPosBug` could therefore be executed at some point. With the HB relation we can directly determine this and create a mutation, where `do 2` is successful, making it possible to find the potential bug with just two runs.

GoPie only reorders channel and mutex operations. It is therefore not
able to rewrite the schedule in such a way, that the `codeWithPosBug` is triggered (unless it is accidentally created by some mutation).
But even vor GoPie+, which includes all operations, this approach can be beneficial.

To see the general advantage of out approach, assume it could reschedule the `Do`.\
GoPie will always create scheduling chains such that two neighboring operations in this chain are in different routines and then randomly
modifies those chains, by abridging, flipping, substituting or augmenting the chain. There are a huge number of possible scheduling chains that can be generated, only a very small number of them would contains the potential bug.
GoPie would therefore need to run a large number of unnecessary runs , hoping that one of them leads to the detection of the potential bug.
If we also include that GoPie would need to trigger the bug to
detect it, meaning not all chains that execute `codeWithPosBug` may result in the detection of the bug, the number of required runs increase even more, while we are able detect the bug in only two runs, assuming the bug can be detected with the HB analysis when running `codeWithoutPosBug`.

The same type of scenario can be constructed for each of the proposed flow change fuzzing patterns (once, chan order, try-lock).