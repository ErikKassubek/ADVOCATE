# Atomic
The recording of atomics records atomic operations, both on normal types and on atomic types. This includes Add, CompareAndSwap, Swap, Load and Store operations.


## Trace element:
The basic form of the trace element is
```
A,[tpost],[id],[opA],[pos]
```
where `A` identifies the element as an atomic operation.
The other fields are set as follows:
- [tpost]: This field shows the value of the internal counter when the operation is executed.
- [id]: This field shows a number representing the variable. It is not possible to give every variable its own unique, consecutive id. For this reason, this id is equal to the memory position of the variable.
- [opA]: This field shows the type of operation. Those can be
	- `L`: Load
	- `S`: Store
	- `A`: Add
	- `W`: Swap
	- `C`: CompareAndSwap
	- `U`: unknown (should not appear)
- [pos]: The last field show the position in the code, where the atomic operation was executed


## Implementation
TODO: write implementation