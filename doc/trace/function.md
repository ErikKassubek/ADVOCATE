# Function

Both the start of a function and the return of a function are recorded in the trace.
As an example, see the following.

```go
def f(){
    ...
}

func main() {  // routine 1, line 1
    f()
}
```

Both the call of function `f` and the return of function `f` are recorded.

## Trace element

This will create 2 trace elements.

The call of a function is recorded as

```
F,[t],[posCall],[posFunc]
```

where `F` identifies the element as a function call.

- [t] $\in \mathbb N$: This is the time. It is replaced by the int value of the global counter at the moment of the routines creation.
- [posCal]: Position of the function call (f())
- [posFunc]: Position of the function definition (def f(){...)

The return of a funciton is recorded as 

```
R,[t]
```

where `R` identifies the element as a function call.

- [t] $\in \mathbb N$: This is the time. It is replaced by the int value of the global counter at the moment of the routines creation.

The element does not contain a position. The function corresponding to the return is identified by the last preceding call element that has not been returned (FILO).

## Implementation

TODO: call

The return element is recorded in `deferreturn` [deferreturn](../../goPatch/src/runtime/panic.go#L579). This function is automatically called when a function returnes, with the main task of the function beeing to call the `defer` in the function. The recording of the return element is added after all the defered functions have been called.
