# WaitGroup
The Add, Done and Wait operations of a wait group are recorded in the trace where the operations occurs.

## Trace element
The basic form of the trace element is 
```
W,[tpre],[tpost],[id],[opW],[delta],[val],[pos]
```
where `W` identifies the element as a wait group element. The following sets are set as follows:

- [tpre] $\in\mathbb N: This is the value of the global counter when the operation starts
the execution of the lock or unlock function
- [tpost] $\in\mathbb N: This is the value of the global counter when the operation has finished
its operation.
- [id] $\in\mathbb N: This is the unique id identifying this wait group
- [opW]: This filed identifies the operation type that was executed on the wait group:
    - [opW] = `A`: change of the internal counter by delta. This is done by Add or Done.
    - [opW] = `W`: wait on the wait group
- [delta]$\in \mathbb Z$ : This field shows the change of the internal value of the wait group.
For Add this is a positive number. For Done this is `-1`. For Wait this is always 
`0`.
- [val]$\in \mathbb N_0$ : This field shows the new value of the internal counter after the operation 
finished. This value is always greater or equal 0. For Wait, this field must be `0`.
- [pos]: The last field show the position in the code, where the mutex operation 
was executed. It consists of the file and line number separated by a colon (:)

## Example
The following is an example program for a wait group
```go
package main

import (
    "sync"
)

func main() {  // routine 1
    var wg sync.WaitGroup

    wg.Add(1)

    go func() {  // routine 2
        ...
        wg.Done()
    }
    
    wg.Wait()
}
```
If we ignore all internal operations we get the following trace:
```txt
W,1,2,1,A,1,1,example_file.go:10;G,3,2;W,6,7,1,W,0,0,example_file.go:17
```
```txt
W,4,5,1,A,-1,1,example_file.go:14
```
## Implementation
The recording of the operations is done in the `go-patch/src/sync/waitgroup.go` file in the `Add` (Add, Done) and `Wait` functions. To save the id of the wait group, an additional 
field is added to the `WaitGroup` struct.