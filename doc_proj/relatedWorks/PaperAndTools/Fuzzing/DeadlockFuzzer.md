# A Randomized Dynamic Program Analysis Technique for Detecting Real Deadlocks

[P. Joshi, C.-S. Park, K. Sen, and M. Naik, “A randomized dynamic program analysis
technique for detecting real deadlocks,” SIGPLAN Not., vol. 44, no. 6, pp. 110–120, Jun.
2009, issn: 0362-1340. doi: 10.1145/1543135.1542489. [Online]. Available: https://doi.org/10.1145/1543135.1542489](https://doi.org/10.1145/1543135.1542489)

## Summary

The developed tool of this paper is called DeadlockFuzzer. The goal is the detect deadlocks in multi-threaded java program.
It uses a record-analyze-replay approach.

The technique runs in two stages.
In the first stage, it uses an imprecise dynamic analysis technique
to find potential deadlocks in a multi-threaded program by observing an execution of the program. In the second stage, it controls
a random thread scheduler to create the potential deadlocks with
high probability.

The dynamic analysis is based on constructing lock dependency graphs and finding cycles in this graph while avoiding duplicates.

It then second stage it runs a replay for each found cycle as follows:

```
s ⇐ s0  // initial state
Paused ⇐ ∅
LockSet and Context map each thread to an empty stack

while Enabled(s) ≠ ∅ do
    t ⇐ a random thread in Enabled(s) \ Paused
    Stmt ⇐ next statement to be executed by t

    if Stmt = c : Acquire(l) then
        push l to LockSet[t]
        push c to Context[t]
        checkRealDeadlock(LockSet) // see Algorithm 4

        if (( abs(t), abs(l), Context[t] ) ∉ Cycle) then
            s ⇐ Execute(s,t)
        else
            pop from LockSet[t]
            pop from Context[t]
            add t to Paused
        end if

    else if Stmt = c : Release(l) then
        pop from LockSet[t]
        pop from Context[t]
        s ⇐ Execute(s,t)

    else
        s ⇐ Execute(s,t)
    end if

    if |Paused| = |Enabled(s)| then
        remove a random thread from Paused
    end if

end while

if Active(s) ≠ ∅ then
    print ‘System Stall!’
end if
```