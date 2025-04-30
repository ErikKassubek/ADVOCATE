
# Cyclic Deadlock


A cyclic deadlock occurs, if locks from multiple routines block each other in
a cyclic manor. We use lock trees to find such circles. We additionally use the HB analysis,
to see, if see, if the operations in the cycle are concurrent.

Lets, as an example assume, the program trace with the mutexes m, n and o had
the following form:

```
  T1         T2          T3
lock(m)
unlock(m)
lock(m)
lock(n)
unlock(m)
unlock(n)
           lock(n)
           lock(o)
           unlock(o)
           unlock(n)
                       lock(o)
                       lock(m)
                       unlock(m)
                       unlock(o)
```

Whis would result in the following lock trees:

```
 T1   T2  T3
  m   n   o
  |   |   |
  n   o   m
```

which can be connected to a circle as follows:

```
 T1   T2  T3
  m   n   o
/ | / | / |
| n   o   m
\--------/
```
