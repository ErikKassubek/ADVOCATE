# Recording

To analyze a program we have to record the relevant operations. Which operations are recorded is described in `trace.md` and the `traceElements` directory.

## Modified Go

To record this trace, we use a modified version of the Go runtime and standard library, that includes functionality to record these elements. The modified version can be found in the `go-patch` directory. We have to build the new compiler and runtime. This can be done my navigating into `go-patch/src` and running the `all.bash` or `all.bat` file. Potentially failing tests can be ignored. This will create a `go-patch/bin` directory containing a `./go` file. This file can be used like the normal `go` command to build or run programs (e.g. `./go build` or `./go run main.go`). T make it work we also have to change the `GOROOT` environment variable to point to the new runtime by running e.g.

```bash
export GOROOT=$HOME/ADVOCATE/go-patch/
```

## Recording

To record the program, we have to add a header at the beginning of the main function:

```go
advocate.InitTracing(0)
defer advocate.Finish()
```

It has to be included before any other code. It is also necessary to import the `advocate` library. 

We now run the program like normal (with the created `./go` program in `go-patch/bin`). The trace files will be automatically created. It will be created in the folder `advocateTrace`.

## Known problems

### Holding Locks

In some cases this will result in an

```
fatal error: schedule: holding locks
```

error. In this case increase the argument in `InitTracing` until
the problem disappears.

In some cases, the trace files can get very big. If you want to simplify the 
traces, you can set the value in `InitAtomics` to $-1$. In this case, 
atomic variable operations are not recorded. 

### Panic

If the program is stopped because of an panic while the main routine is
sleeping, it is possible, that no trace files or only partial trace files are
created. If the panic
occurs only occasional, it is possible to run the program again, to get
the trace file (assuming the panic does not occur again). If it occurs
in every run, it is necessary to first fix this panic, before a trace can
be recorded.

## Implementation
The recording is implemented by patching the go runtime, meaning a recording 
function is added into the recorded function, which records the operation.
The implementation for the different recorded types are explained in the `traceElemens`
folder. The different routines are all recorded individually, by adding a trace 
object into the `g` object (defined in `runtime/runtime2.go`), which is created 
automatically for each routine by the go runtime. Additionally we use a global 
counter. For each operation we store this unique counter when the operation
has finished, and for some additionally when whey started. With this global counter 
it is possible to create one global trace from the different local traces.

At the end all traces are written into individual files. Storing the full trace 
for a program internally can lead to the situations where the computer does not 
have enough RAM. To prevent this, we run an internal go routine to monitor the 
free RAM. In this case, we pause the execution of the program, write the current 
traces to the files and delete the internal trace. We additionally force the 
garbage collector to run. After that we can continue the execution of the 
program.