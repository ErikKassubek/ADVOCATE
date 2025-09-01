# Runtime

We want to be able to record the execution of a go program and to
replay a given trace. In many similar works, this is done by directly
annotating the code of the program that should be analyzed or replayed.
We want to avoid this. We therefore directly integrate the [recording](recording.md)
and [replay](replay.md) creating an modified go runtime.

The runtime can be found in the [goPatch](../goPatch/) folder.
It is currently based on go1.24.1

Before we can use this runtime, we first need to build the runtime.
This can be done by running [this](../goPatch/src/make.bash) script,
which will create the runtime in the [go-path/bin](../goPatch/bin/).

If we want to use this runtime directly, we need to change the
`GOROOT` environment variable to this go-path file, e.g. with

```shell
export GOROOT=/home/.../goPatch
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

- [src/advocate/advocate_fuzzing.go](../goPatch/src/advocate/advocate_fuzzing.go)
- [src/advocate/advocate_replay.go](../goPatch/src/advocate/advocate_replay.go)
- [src/advocate/advocate_tracing.go](../goPatch/src/advocate/advocate_tracing.go)
- [src/runtime/advocate_exit.go](../goPatch/src/runtime/advocate_exit.go)
- [src/runtime/advocate_fuzzing.go](../goPatch/src/runtime/advocate_fuzzing.go)
- [src/runtime/advocate_ids.go](../goPatch/src/runtime/advocate_ids.go)
- [src/runtime/advocate_replay.go](../goPatch/src/runtime/advocate_replay.go)
- [src/runtime/advocate_routine.go](../goPatch/src/runtime/advocate_routine.go)
- [src/runtime/advocate_time.go](../goPatch/src/runtime/advocate_time.go)
- [src/runtime/advocate_trace.go](../goPatch/src/runtime/advocate_trace.go)
- [src/runtime/advocate_trace_atomic.go](../goPatch/src/runtime/advocate_trace_atomic.go)
- [src/runtime/advocate_trace_channel.go](../goPatch/src/runtime/advocate_trace_channel.go)
- [src/runtime/advocate_trace_cond.go](../goPatch/src/runtime/advocate_trace_cond.go)
- [src/runtime/advocate_trace_mutex.go](../goPatch/src/runtime/advocate_trace_mutex.go)
- [src/runtime/advocate_new_elem.go](../goPatch/src/runtime/advocate_trace_new_elem.go)
- [src/runtime/advocate_trace_once.go](../goPatch/src/runtime/advocate_trace_once.go)
- [src/runtime/advocate_trace_routine.go](../goPatch/src/runtime/advocate_trace_routine.go)
- [src/runtime/advocate_trace_select.go](../goPatch/src/runtime/advocate_trace_select.go)
- [src/runtime/advocate_trace_waitgroup.go](../goPatch/src/runtime/advocate_trace_waitgroup.go)
- [src/runtime/advocate_tracing.go](../goPatch/src/runtime/advocate_tracing.go)
- [src/runtime/advocate_util.go](../goPatch/src/runtime/advocate_util.go)
- [src/runtime/advocate_wait.go](../goPatch/src/runtime/advocate_wait.go)

### Modified files

Modifications in files are marked with

```go
// ADVOCATE-START
...
// ADVOCATE-END
```

- [src/runtime/proc.go](../goPatch/src/runtime/proc.go)
- [src/runtime/runtime2.go](../goPatch/src/runtime/runtime2.go)
- [src/runtime/chan.go](../goPatch/src/runtime/chan.go)
- [src/runtime/select.go](../goPatch/src/runtime/select.go)
- [src/runtime/panic.go](../goPatch/src/runtime/panic.go)
- [src/sync/cond.go](../goPatch/src/sync/cond.go)
- [src/sync/mutex.go](../goPatch/src/sync/mutex.go)
- [src/sync/rwmutex.go](../goPatch/src/sync/rwmutex.go)
- [src/sync/once.go](../goPatch/src/sync/once.go)
- [src/sync/waitgroup.go](../goPatch/src/sync/waitgroup.go)
- [src/sync/atomic/asm.s](../goPatch/src/sync/atomic/asm.s)
- [src/sync/atomic/doc_32.go](../goPatch/src/sync/atomic/doc_32.go)
- [src/sync/atomic/doc_64.go](../goPatch/src/sync/atomic/doc_64.go)
- [src/sync/atomic/doc.go](../goPatch/src/sync/atomic/doc.go)
- [src/sync/atomic/type.go](../goPatch/src/sync/atomic/type.go)
- [src/cmd/compile/internal/ssagen/intrinsics.go](../goPatch/src/cmd/compile/internal/ssagen/intrinsics.go)
- [src/cmd/link/internal/loader/loader.go](../goPatch/src/cmd/link/internal/loader/loader.go)
