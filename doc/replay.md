# Replay

Replay allows us to force the execution of a program to follow a given trace.

## Toolchain

The replay is run, if the test of main function starts with the following header:

  ```go
  advocate.InitReplay(index, exitCode, timeout, atomic)
  defer advocate.FinishReplay()
  ```

The parameters are as follows:

- `index` $\in \mathbb{N}$: The replay expects the name of the folder containing the trace to be called `advocateTrace`, in this case set `index = 0` or `rewritten_trace_[index]`, meaning if the trace folder is called `rewritten_trace_2`, set `index = 2`
- `exitCode` $\in \mathbb B$: currently always `false`<!-- TODO: currently always set to false, maybe remove -->
- `timeout` $\in \mathbb{N}$: If you want to cancel the replay with a timeout after a given time, set this value to the timeout in seconds. Otherwise set to `0`.
- `atomic`: If set to `true`, the replay will force the correct order of atomic events. If atomic operations should be ignored for the replay, set this to `false`.

When using the toolchain to run replays, this header is automatically added.



## Implementation
The following is a description of the implementation of the trace replay.

The code for the replay is mainly in [advocate/advocate_replay.go](../go-patch/src/advocate/advocate_replay.go) and [runtime/advocate_replay.go](../go-patch/src/runtime/advocate_replay.go) as well as in the code implementation of all recorded operations.

[advocate/advocate_replay.go](../go-patch/src/advocate/advocate_replay.go) mainly contains the code to read in the trace and initialize the replay. When reading in the trace, all trace files are read. The internal representation of the replay consists of a `map[int][]ReplayElement` `replayData`. For each of the routines, we store a slice
of `ReplayElement`, meaning the list of elements in this routine.
The lists are sorted by the `tPost` time stamp. Each `ReplayElement`
represents one operation that is to be executed. It contains data about
the type of operation, the timestamp and the position of the operation
in the file. This position is used to connect the `ReplayElement` to the actual operation during the execution. Additionally it may contain
information about wether the operation is blocked, meaning it should start but never finish or wether it should execute successfully, e.g. for `once.Do` and `TryLock`. For selects, it also contains the internal
index of the case that should be executed.

[runtime/advocate_replay.go](../go-patch/src/runtime/advocate_replay.go) contains the functions for the actual order enforcement.\
The basic idea is as follows: When a operation wants to
execute, it informs the replay manager that it wants to execute. It will then wait until the manager clears it to run. The manager will clear the operations in the order
in which they appear in the trace. In most cases, the manager will then wait
for an acknowledgement from the operation, that it has finished executing.
For channel send/recv we do not wait for an acknowledgement, since here multiple
operation need to be executed at the same time.
To prevent the replay from getting stuck, the manager is able to release operations out of order if it thinks, something went wrong.\
The Order enforcements consists of the operation and the replay manager. First, we have the code in operations itself, that blocks the execution of the operation, until its released by the replay manager.\
The manager will release the operations in the correct order.

### Replay in operations
![Replay in Operations](img/replayInOp.png)\
When a operation wants to execute, it will call the `WaitForReplay` function. The arguments of the function
contain information about the waiting operation (type of operation and
skip value for `runtime.Caller`) as well as information about wether the
operation will send an acknowledgement (`wAck`) or not. The function returns a
`wait` boolean and to channels `ch` and `ack`. The `wait` value tells the
operation wether it needs to wait. It if false, e.g if the replay is disabled
or if the operation is ignored (we ignore all internal operations). If `wait = false`, the operation will just continue to execute normally. Otherwise,
it will start to read on the the `ch` channel. When the manager clears the
operation to run, it will send a message over this channel. This
message, contains information about wether the operation blocked and wether
it was successful. If it should block (tPost = 0), it will block the operation
forever. If the operation was not successful (only possible for once.Do or Try(R)Lock), it will force the execution of the operation to follow this behavior.
Otherwise it will now execute the operation. When `wAck` is true, the
operation will send an empty message as soon as the operation is finished
(normally implemented by `defer`). This allows the manager to release the
next operation.


## Replay Manager
![Replay Manager](img/replayManager.png)\
The replay manager releases the operations in the correct order.

When an
operations wants to execute, it calls the `WaitForReplay` functions. If the
replay is not enable or if the operation is an internal operation and is
therefore ignored, the function will just return. Otherwise it will
create a wait channel `chWait` and and acknowledgement channel `chAck`.
It will store those channels with a reference to the code location and routine
id of the operation. The routine id is the id of the routine in the replay trace. It is set for the routine in the `newProc` function in `runtime/proc.go`. This allows us to separate operations in the same code position but in separate routines. Code at the same code position and in the same routine does not need
to be separated, since routines are executed sequentially. The function then
returns the channel to the waiting operation.

To release the operations, a separate routine `ReleaseWait` is run in the
background. CONT
