# Runtime

We want to be able to record the execution of a go program and to
replay a given trace. In many similar works, this is done by directly
annotating the code of the program that should be analyzed or replayed.
We want to avoid this. We therefore directly integrate the [recording](recording.md)
and [replay](replay.md) creating an modified go runtime.

The runtime can be found in the [go-patch](../go-patch/) folder.
It is currently based on go1.24.1

Before we can use this runtime, we first need to build the runtime.
This can be done by running [this](../go-patch/src/make.bash) script,
which will create the runtime in the [go-path/bin](../go-patch/bin/).

If we want to use this runtime directly, we need to change the
`GOROOT` environment variable to this go-path file, e.g. with

```shell
export GOROOT=/home/.../go-patch
```

If the toolchain is used, this is done automatically.

The only change required in the analyzed program is to add
a header at the main or test function as described in the [recording](recording.md)
and [replay](replay.md).\
If the toolchain is used, this is done automatically.

## Changed files

The following is a list of all files in the runtime that have been added or modified.\
All modifications have been annotated with ADVOCATE-START and ADVOCATE-END.

### Added files

- src/advocate/advocate_fuzzing.go
- src/advocate/advocate_replay.go
- src/advocate/advocate_tracing.go
- src/runtime/advocate_fuzzing.go
- src/runtime/advocate_replay.go
- src/runtime/advocate_routine.go
- src/runtime/advocate_trace_atomic.go
- src/runtime/advocate_trace_channel.go
- src/runtime/advocate_trace_cond.go
- src/runtime/advocate_trace_mutex.go
- src/runtime/advocate_trace_once.go
- src/runtime/advocate_trace_routine.go
- src/runtime/advocate_trace_select.go
- src/runtime/advocate_trace_waitgroup.go
- src/runtime/advocate_trace.go
- src/runtime/advocate_util.go

### Modified files

- src/runtime/proc.go
- src/runtime/runtime2.go
- src/runtime/chan.go
- src/runtime/select.go
- src/runtime/panic.go
- src/sync/cond.go
- src/sync/mutex.go
- src/sync/rwmutex.go
- src/sync/once.go
- src/sync/waitgroup.go
- src/sync/atomic/advocate_atomic.go
- src/sync/atomic/asm.s
- src/sync/atomic/doc.go
- src/sync/atomic/doc_32.go
- src/sync/atomic/doc_64.go
- src/sync/atomic/type.go
- src/cmd/compile/internal/ssagen/intrinsics.go
- src/cmd/linl/internal/loader/loader.go