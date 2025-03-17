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

### Flow
#### Replay in operations
<img src="img/replayInOp.png" alt="Replay in Operations" width="600px" height=auto>
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


#### Replay Manager
![Replay Manager](img/replayManager.png)\
The replay manager releases the operations in the correct order.

When an
operations wants to execute, it calls the `WaitForReplay` functions. If the
replay is not enable or if the operation is an internal operation and is
therefore ignored, the function will just return. Otherwise it will
create a wait channel `chWait` and and acknowledgement channel `chAck`.
It will store those channels with a reference to the code location and routine
id of the operation. The routine id is the id of the routine in the replay trace. It is set for the routine in the [newProc](../go-patch/src/runtime/proc.go#L5057) function in `runtime/proc.go`. This allows us to separate operations in the same code position but in separate routines. Code at the same code position and in the same routine does not need
to be separated, since routines are executed sequentially. The function then
returns the channel to the waiting operation.

To release the operations, a separate routine `ReleaseWait` is run in the
background.\
The main loop of this manager is as follows:
First, the manager gets the next element that should be executed (see [here](#getnextreplayelement)). If no element is left, the replay is disabled.
The same is true if the next element in the trace is the `replayEnd` element.
If the bug is confirmed by reaching a certain point in the code (e.g leak or
resource deadlock), this will confirm the replay. If the element is an
operation element, the manager will check if the element should be ignored
for the replay. In this case, there is no waiting operation and the
manager will simply assume, that the operation was executed and will continue
the loop from the beginning. If this is not the case, the it is checked, if the
operation is already waiting. If this is the case, it is [released](#release).
After that, the manager will continue the loop to get the next trace element.
If the operation is not yet waiting, the manager will directly restart the
loop without advancing to the next replay element, until the operation is
waiting.

It is possible, that in the replay or the rewrite something went wrong
and there are a small number of trace elements, that cause the replay to get
stuck. To still get a chance that is may resolve itself, the manager
is able to release the longest waiting operation, if it senses that the
replay is stuck. This will be done, if the next replayElement is the same for
a too long time. Additionally, if no element has been cleared regularly
for a certain time, the replay will assume that it is stuck and cannot be
brought back by releasing the oldest waiting elements and will therefore
disable the replay completely, meaning the program will continue without
any guidance (all waiting operations are released). If the replay is already
far enough, this may still result in the program running in the expected bug.
For more information see the bottom left corner of the flow diagram.

#### getNextReplayElement
The trace is stored in a map. Each entry contains the elements for one routine as a sorted list. Additionally, we have a map `replayIndex` with the same key. The values of this map contain for each routine the index of the first element in the trace list, that has not been executed yet. We now iterate over all routines that have elements that have not been executed yet. For each
of those traces we check the first element that has not been
executed yet and choose the one with the smallest time value
as the next element to be executed.\
When the replay manager releases an oldest element, we need to
make sure, that when this element is returned as the
next element to be replayed, we skip it, since it already has been executed. We therefore have an map `alreadyExecutedAsOldest` from the element key to a counter. This counter contains how often the element has been executed without it being the next element. If this value for the
next element to be replayed is not 0, we will advance the `replayIndex` for the routine, reduce the value of `alreadyExecutedAsOldest` of the element by 1 and call the `getNextReplayElement` function again, to get the next element to be executed.


#### release
To release a waiting element, we send the element info over the corresponding
channel, on which the operation is waiting. If an acknowledgement is expected,
the release function will then read on the acknowledgement channel until
the acknowledgement is received. To prevent a failed acknowledgement from
getting the replay stuck, we have a timeout. If no acknowledgement has been received
after a certain time, we continue the replay anyways. To prevent the case,
where an operation ties to send an acknowledgement after the timeout has been triggered
and therefore cannot send, we set the buffer size of all acknowledgement
channels to 1. They can therefore also send if no one is receiving any more.

#### foundReplayElement/releasedElementOldest
If a replay has been released, the `replayIndex` value of the corresponding
routine is advanced by one. If an oldest replay element was released without
it being the next element, the `alreadyExecutedAsOldes` counter for this element
is increased by one.
![Release](img/release.png)


### Select
While most operations are fully determined by the order in which they are executed,
this may not be the case for selects. Here we need to determine which case
the replay should execute. For this an alternative implementation of the
[select structure with a preferred case](../go-patch/src/runtime/select.go#L151) has been implemented.

In the [original implementation](../go-patch/src/runtime/select.go#L629), the select has three phases. In the
first phase the code iterates over all cases and checks, if one of then can
immediately be executed.\
If this is the case, it is executed and the select returns. If this is not
the case, it is checked if there is a default case. If there is, it is executed.
If not, the select start its second phase. Here all channels in all the cases
enqueue the corresponding operations. The routine then parks, until
it is woken up by some routine that wants to communicate with one of the channels.
The communication is then executed. In the third phase, all communications
that have not been executed are dequeued again.

For the select with preferred case, the following changes have been made.
In the first pass we only check the case, where the internal index `casi`
is equal to the preferred index. If the default case is the preferred index,
we don't do anything here.\
If the preferred case is the default case, we directly execute the default case.\
In the second pass, we only enqueue the preferred case.

Similar to release of the oldest waiting element, we also want to release the
wait on the select if the runtime senses, that the replay may be stuck. For
this reason, we implement a `gopark` function with a timeout, that
automatically wakes it up after a certain time. The park with timeout is implemented
in [prog.go](../go-patch/src/runtime/proc.go#L439) as follows:
```go
func goparkWithTimeout(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceReason traceBlockReason, traceskip int, timeout int64) {
	mp := acquirem()
	gp := mp.curg

	// Setup timer if timeout is non-zero
	if timeout > 0 {
		go func() {
			sleep(float64(timeout))
			if readgstatus(gp) == _Gwaiting {
				goready(gp, traceskip)
			}
		}()
	}

	// original gopark implementation
  ...
}
```
The additionally routine will wake the sleeping routine if the timer
runs out before it is awoken by a communication partner.

When the routine is woken up by a communication partner, we do not need to
pass the third phase, since there are no more waiting channel in the select.
In this case the select returns. If it was awoken by the timeout, we dequeue
the waiting channel operation and return. In this case, the select is run again,
now with the original implementation, meaning without any guidance.

Selects with only one case and a default are implemented separately in go ([here](../go-patch/src/runtime/chan.go#L990) and [here](../go-patch/src/runtime/chan.go#L1057)).
This implementation tries to directly [send](../go-patch/src/runtime/chan.go#L201) or
[receive](../go-patch/src/runtime/chan.go#L645) on the channel in the non default case.
If this is not possible because no communication partner is available, it directly
calls the default.\
For the replay this is changed as follows: If the default case is the preferred case,
we directly run the default case without trying to send/recv. If the non-default
case is the preferred one, we busy-wait with a for loop, each time trying to send
or receive, until the communication succeeds or a timer runs out. If the timer
runs out first, we run the default case regardless.

### Atomics
While for most operations, we could add the code for the [replay](#replay-in-operations)
directly into the implementation of those operations, this was not
possible for the atomic operations, since they are partially implemented in
assembly. We therefore needed to intersect an additional function call.
For more information about this, see the [atomic recording documentation](./trace/atomics.md#implementation).
