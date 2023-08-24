# Trace

The following is the structure of the trace T in EBNF. There is an extra, 
better readable
explanation for all trace elements in the corresponding files in `traceElements`. 
```
T := L\nT | ""                                                  (trace)
L := "" | {E";"}E                                               (routine local trace)
E := G | M | W | C | S                                          (trace element)
G := "G,"tpre","id                                              (element for creation of new routine)
A := "A,"tpre","addr","opA                                      (element for atomic operation)
M := "M,"tpre","tpost","id","rw","opM","suc","pos               (element for operation on sync (rw)mutex)
W := "W,"tpre","tpost","id","opW","delta","val","pos            (element for operation on sync wait group)
C := "C,"tpre","tpost","id","opC","cl",oId","qSize","qCountPre","qCoundPost","pos             (element for operation on channel)
S := "S,"tpre","tpost","id","cases","chosen","oId","pos         (element for select)
tpre := ℕ                                                       (timer when the operation is started)
tpost := ℕ                                                      (timer when the operation has finished)
addr := ℕ                                                       (pointer to the atomic variable, used as id)
opA := "L" | "S" | "A" | "W" | "C" | "U"                        (operation type of the atomic operation)
id := ℕ                                                         (unique id of the underling object)
rw := "R" | "-"                                                 ("R" if the mutex is an RW mutex, "-" otherwise)
opM := "L" | "R" | "T" | "Y" | "U" | "N"                        (operation on the mutex, L: lock, R: rLock, T: tryLock, Y: tryRLock, U: unlock, N: rUnlock)
opW := "A" | "W"                                                (operation on the wait group, A: add (delta > 0) or done (delta < 0), W: wait)
opC := "S" | "R" | "C"                                          (operation on the channel, S: send, R: receive, C: close)
suc := "t" | "f"                                                (the mutex lock was successful ("t") or it failed ("f", only possible for try(r)lock))
cl := "t" | "f"                                                 (If this value is set to `t`, the operation was finished, because the channel was closed in another routine, while or before the channel was waiting at this operation.)
pos := file":"line                                              (position in the code, where the operation was executed)
file := 𝕊                                                       (file path of pos)
line := ℕ                                                       (line number of pos)
delta := ℕ                                                      (change of the internal counter of wait group, normally +1 for add, -1 for done)
val := ℕ                                                        (internal counter of the wait group after the operation)
oId := ℕ                                                        (identifier for an communication on the channel, the send and receive (or select) that have communicated share the same oId)
qSize := ℕ                                                      (size of the channel queue, 0 for unbufferd)
qCountPre := ℕ                                                  (number of elements in the queue before the operation)
qCountPost := ℕ                                                 (number of elements in the queue after the operation)
cases := case | {case"."}case                                   (list of cases in select, seperated by .)
case := cId""("r" | "s") | "d"                                  (case in select, consisting of channel id and "r" for receive or "s" for send. "d" shows an existing default case)  
cId := ℕ                                                        (id of channel in select case)
chosen := ℕ0 | "-1"                                             (index of the chosen case in cases, -1 for default case)    
```

The trace is stored in one file. Each line in the trace file corresponds to one 
routine in the executed program. The line number is equal to the id of the routine.
The elements in each line are separated by 
semicolons (;). The different fields in each element are seperated by 
commas (,). The first field always shows the type of the element:

- G: creation of a new routine
- A: atomic operation
- M: mutex operation
- W: wait group operation
- C: channel operation
- S: select operation

The other fields are explained in the corresponding files in the `traceElements` directory.
These files also describe how the trace elements are recorded.

## Implementation
The runtime of Go creates a struct `g` for each routine (implemented in `go-patch/src/runtime/runtime2.go`). This routine is used to locally store the trace for each routine. 
In it, an additional field is added, storing the id of the routine, a reference to `g` and the list of trace elements (`Trace`) recorded for this routine. When creating a new routine, this list is created. A reference to this list is additionally stored in a map called `DedegoRoutines`, to prevent if from being deleted by the garbage collector.

To record the exact temporal schedule of the program, a global counter is added, 
that is always implemented when the tace is changed. This counter is used as 
a timer in the trace.

In the runtime package, it is possible to get the `g` for the currently run routine. If an element that is supposed to be recorded happens, the routine grabs the `g` of the routine where it happens, and adds the new element to the Trace stored in this `g`. The implementation of the functions, that add the new elements in the trace can be found in `go-patch/src/runtime/dedego_trace.go` with additional functions in `go-patch/src/runtime/dedego_routine.go` and `go-patch/src/runtime/dedego_util.go`. The functions defined in `dedego_trace.go`, are called in the functions where the operations on Mutexes, Channels and so on are defined, to record the executions of those operations. The implementation of those functions are additionally described in the files of the respective elements in the traceElements folder.

After the program is finished, the Traces of all routines with references in `DedegoRoutines` are written into a single trace file by the header, that was 
added to the program.