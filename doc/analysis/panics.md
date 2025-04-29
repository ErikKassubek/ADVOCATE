# Panics

We detect occurring panics directly in the runtime. For this, we add a function
to the implementation of the [panic](../../go-patch/src/runtime/panic.go#L744) function.

This panic is then always automatically called if a program panics.

Based on the type and content of the  message in the panic, we can sort the
panics based on our interests into

- "send on closed channel"
- "close of closed channel"
- "close of nil channel"
- "sync: negative WaitGroup counter"
- "test timed out"
- "sync: RUnlock of unlocked RWMutex"
- "sync: Unlock of unlocked RWMutex"
- "sync: unlock of unlocked mutex"
- other

This info can then used during the replay to determine if an expected bug
has been triggered and for the recording and fuzzing to directly
find actual bugs.