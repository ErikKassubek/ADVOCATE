# Changed files

The following is a list of all files in the Go runtime that have been
added or changed.

Added files:

- src/runtime/advocate_routine.go
- src/runtime/advocate_trace.go
- src/runtime/advocate_trace_atomic.go
- src/runtime/advocate_trace_channel.go
- src/runtime/advocate_trace_cond.go
- src/runtime/advocate_trace_mutex.go
- src/runtime/advocate_trace_routine.go
- src/runtime/advocate_trace_select.go
- src/runtime/advocate_trace_waitgroup.go
- src/runtime/advocate_util.go
- src/runtime/advocate_replay.go
- src/runtime/advocate_time.go
- src/advocate/advocate_fuzzing.go
- src/advocate/advocate_replay.go
- src/advocate/advocate_tracing.go
- src/sync/atomic/advocate_atomic.go

Changed files (marked with ADVOCATE-CHANGE):

- src/runtime/proc.go
- src/runtime/runtime2.go
- src/runtime/chan.go
- src/runtime/select.go
- src/runtime/panic.go
- src/sync/atomic/doc.go
- src/sync/atomic/type.go
- src/sync/atomic/asm.s
- src/sync/mutex.go
- src/sync/rwmutex.go
- src/sync/waitgroup.go
- src/sync/once.go
- src/sync/cond.go
- src/sync/pool.go
- src/internal/poll/fd_poll_runtime.go
- cmd/compile/internal/ssagen/ssa.go


Additionally some test files have been altered.