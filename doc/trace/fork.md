# New Routine
The spawning of a new routine is recorded in the trace. The following is
and example of code where such an trace element is recorded.
```go
func main() {  // routine 1, line 1
    go func() {  // routine 2, line 2
        ...
    }
}
```
In the main routine (routine 1) a new routine (routine 2) is spawned using the
`go` keyword.
This is recorded in the trace of routine 1.

## Trace element
This will create 2 trace elements.

In the routine, where the new routine is created, the following element is added.
```
G,[tpost],[id],[pos]
```
where `G` identifies the element as an routine creation element.\
- [tpost] $\in \mathbb N$: This is the time. It is replaced by the int value of the global counter at the moment of the routines creation.
- [id] $\in \mathbb N$: This is the id of the newly created routine. This integer id corresponds with
the line number, where the trace of this new routine is saved in the trace.
- [pos]: Position in the program, where the spawn was created.

If we ignore all other internal elements regarding the counter, the element for
the given example would be stored in the trace as
```txt
G,1,2,.../main.go:2
```
meaning, in routine 1 a new routine with id 2 was created at time 1. The path here is shortened for readability. The actual trace file contains the whole path.


## Implementation
The element is recorded in the `newproc` function in the `go-patch/src/runtime/proc.go` file. Unfortunately it is not possible to record where in the program
files the `go func` command is, because the compiler turns a `go` statement into a call of `newproc` which does contain the information where in the program
file the `go` statement is located.