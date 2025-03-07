# Conditional
The three operations on conditional variables are
- Wait
- Signal
- Broadcast

## Trace element
The basic form of the trace element is
```
D,[tpre],[tpost],[id],[opN],[pos]
```
where `D` identifies the element as a conditional element.
The other fields are set as follows:
- [tpre] $\in \mathbb N$: This is the value of the global counter when the operation starts
the execution of the lock or unlock function
- [tpost] $\in \mathbb N$: This is the value of the global counter when the mutex has finished its operation. For lock operations this can be either if the lock was successfully acquired or if the routines continues its execution without
acquiring the lock in case of a trylock.
- [id] $\in \mathbb N$: This is the unique id identifying this mutex
- [opN]: This field shows the operation of the element. Those can be
  - [opM] = `W`: Wait
  - [opM] = `S`: Signal
  - [opM] = `B`: Broadcast
- [pos]: The last field show the position in the code, where the mutex operation
was executed. It consists of the file and line number separated by a colon (:)

## Implementation
The recording of the mutex operations is implemented in the `go-patch/src/sync/cond.go` file in the implementation of the `Wait`, `Signal` und `Broadcast` functions.\
To save the id of the conditional, a field for the id is added to the `Cond` struct.\
The recording consist of two function calls, one at the beginning and one at the end of each function.
The first function call is called before the Operation tries to executed
and records the id ([id]) and called operation (opN), the position of the operation in the program ([pos]) and the counter at the beginning of the operation ([tpre]).\
The second function call records the success of the operation. This includes
the counter at the end of the operation ([tpost]).
The implementation of those function calls can be found in
`go-patch/src/runtime/advocate_trace.go` in the functions `AdvocateCondPre`, and `AdvocateCondPost`.
