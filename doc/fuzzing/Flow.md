# Flow Changing Operations

Our bug detection is based on HB analysis. It is therefore not really useful
in the fuzzing, to create mutations, that lead to the same happens before
relations, since they do not produce new information.

Instead we want to focus on mutation techniques that explore new execution paths.
On of them, that is already implemented is the mutation of the selects,
as based on GoFuzz. Tho following primitive are of interest.

One of those would e.g. be `once`. If we have two `Do` on the same `once` in the
trace, were one of them was executed and the other was not, and both of them
are concurrent based on the HB analysis, one could create a mutation, where
the other once is preferred, either by creating a direct trace to replay or,
which would probably much easier to implement, by creating a run, where the
preciously executed `once.Do` is blocked from actually executing, such that
the other once can be executed. This would allow us to explore the code in the
body of the previously not executed `once.Do`.

Similar scenarios could be created e.g. for `try-locks`.

Another situation we could try to capture, would be for code of the form
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
But to see the general advantage of out approach, assume it could reschedule the `Do`.\
GoPie will always create scheduling chains such that two neighboring operations in this chain are in different routines and then randomly
modifies those chains, by abridging, flipping, substituting or augmenting the chain. I was not able to determine, whether GoPie always create maximal scheduling chain, or whether it looks at different scheduling chains. In both cases we can say, that there are a huge number of possible scheduling chains that can be generated, only a very small number of them would contains the potential bug.
GoPie would therefore need to run a large number of unnecessary runs , hoping that one of them leads to the detection of the potential bug.
If we also include that GoPie would need to trigger the bug to
detect it, meaning not all chains that execute `codeWithPosBug` may result in the detection of the bug, the number of required runs increase even more, while we are able detect the bug in only two runs, assuming the bug can be detected with the HB analysis when running `codeWithoutPosBug`.

The same type of scenario can be constructed for each of the proposed flow change fuzzing patterns (once, chan order, try-lock).