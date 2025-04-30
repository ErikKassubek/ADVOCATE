### Potential send/receive on closed channel

We assume, that the send/recv on the closed channel did not actually occur
in the program run. Let c be the close and a the send or receive operation.
The global trace then has the form:

```
T = T1 ++ [a] ++ T2 ++ [c] ++ T3
```

We now reorder the trace to

```
T = T1 ++ T2' ++ [X_s, c, a, X_e]
```

where T2' = [ g | g in T2 and g <HB c ].\
For send on close, this should lead to a crash of the program. For recv on close, it will probably lead to a different execution of program after the
object. We therefor disable the replay after c and a have been executed and
let the rest of the program run freely. To tell the replay to disable the
replay, by adding a stop character X_e.