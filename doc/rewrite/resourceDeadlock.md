### Cyclick Deadlock
We already get this (ordered) cycle from the analysis (the cycle is ordered in
such a way, that the edges inside a routine always go down). We now have to
reorder in such a way, that for edges from a to b, where a and b are in different
routines, b is run before a. We do this by shifting the timer of all b back,
until it is greater as a.

For the example we therefor get the the following:

```
  T1         T2          T3
lock(m)
unlock(m)
lock(m)
           lock(n)
lock(n)
unlock(m)
unlock(n)
                       lock(o)
           lock(o)     lock(m)
           unlock(o)   unlock(m)
           unlock(n)   unlock(o)
```

If this can lead to operations having the same time stamp. In this case,
we decide arbitrarily, which operation is executed first. (In practice
we set the same timestamp in the rewritten trace and the replay mechanism
will then select one of them arbitrarily).
If this is done for all edges, we remove all unlock operations, which
do not have a lock operation in the circle behind them in the same routine.
After that, we add the start and end marker before the first, and after the
last lock operation in the cycle.
Therefore the final rewritten trace will be

```
  T1         T2          T3
start()
lock(m)
unlock(m)
lock(m)
           lock(n)
lock(n)
                       lock(o)
           lock(o)     lock(m)
end()
```
