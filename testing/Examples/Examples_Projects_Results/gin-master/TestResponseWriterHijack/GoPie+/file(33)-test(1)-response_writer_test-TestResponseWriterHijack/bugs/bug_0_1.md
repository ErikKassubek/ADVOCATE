# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestResponseWriterHijack
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/response_writer_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/ADVOCATE/go-patch/src/runtime/iface.go:102
```go
91 ...
92 
93 	if canfail {
94 		return nil
95 	}
96 	// this can only happen if the conversion
97 	// was already done once using the , ok form
98 	// and we have a cached negative result.
99 	// The cached result doesn't record which
100 	// interface function was missing, so initialize
101 	// the itab again to get the missing function name.
102 	panic(&TypeAssertionError{concrete: typ, asserted: &inter.Type, missingMethod: itabInit(m, false)})           // <-------
103 }
104 
105 // find finds the given interface/type pair in t.
106 // Returns nil if the given interface/type pair isn't present.
107 func (t *itabTableType) find(inter *interfacetype, typ *_type) *itab {
108 	// Implemented using quadratic probing.
109 	// Probe sequence is h(i) = h0 + i*(i+1)/2 mod 2^k.
110 	// We're guaranteed to hit all table entries using this probe sequence.
111 	mask := t.size - 1
112 	h := itabHashFunc(inter, typ) & mask
113 
114 ...
```


## Replay
**Replaying was not run**.

