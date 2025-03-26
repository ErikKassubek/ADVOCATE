# Once

The Do of an Once is recorded the in trace, when where the operation occures.

# Trace element

The basic form of the trace element is

```
O,[tpre],[tpost],[id],[suc],[pos]
```

where `O` identifies the element as a wait group element. The following
fields are

- [tpre] $\in\mathbb N: This is the value of the global counter when the operation starts
  the execution of the lock or unlock function
- [tpost] $\in\mathbb N: This is the value of the global counter when the operation has finished
  its operation.
- [id] $\in\mathbb N: This is the unique id identifying this once
- [suc] $\in \{t, f\}$ records, whether the function in the once was
  executed (`t`) or not (`f`). Exactly on trace element per once must be `t`.
- [pos]: The last field show the position in the code, where the mutex operation
  was executed. It consists of the file and line number separated by a colon (:)

## Example

The following is an example program for a once

```go
package main

import (
    "sync"
)

func main() {
    go doprint1()  // routine 2
    go doprint2()  // routine 3
}

func doprint1() {
    once.Do(func() {  // line 12
        print("Hello, World!")
    })
}

func doprint2() {
	once.Do(func() {  // line 18
		print("Hello, World!")
	})
}
```

If we ignore all internal operations we get the following trace:

```txt
G,1,2;G,2,3;
```
```txt
O,5,6,4,f,/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/bin/main.go:12;
```
```txt
O,3,4,4,t,/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/bin/main.go:18;
```

## Implementation

The recording of the operations is done in the `go-patch/src/sync/once.go` file in the `Do` (Add, Done) and `doSlow` functions. To save the id of the once, an additional field is added to the `Once` struct.
