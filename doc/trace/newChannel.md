# New Channel

This element record the creation of a new channel. It is only used for [fuzzing](../fuzzing/GFuzz.md) and not required for the analysis or replay.

In the future this may be extended to record all new created elements.

# Trace element

The basic form of the trace element is

```
N,[tPost],[id],[elemType],[num],[pos]
```

where `N` identifies the element as a wait group element. The following
fields are

- [tPost] $\in\mathbb N$: This is the value of the global counter when the channel was created
  its operation.
- [id] $\in\mathbb N$: This is the unique id identifying this channel
- [elemType] = "C"
- [num] $\in\mathbb N$: Can be used of additional information. For now this is the channel buffer size.
- [pos]: The last field show the position in the code, where the mutex operation
  was executed. It consists of the file and line number separated by a colon (:)

## Implementation

The main implementation of this is done in the [makechan](../../goPatch/src/runtime/chan.go#L200) function by calling the [AdvocateChanMake](../../goPatch/src/runtime/advocate_trace_new_elem.go#L44) function, that is called if a channel is created with a make. Since a channel can also be created, without a make, the AdvocateChanMake is also called in the send and receive functions if the id of the channel has not been set yet.
