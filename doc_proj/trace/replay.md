# Replay

This element only exists in rewritten traces, not in recorded ones. It signalizes
the start and end of the part of the trace, that was detected as a potential
bug during the analysis and then rewritten. Is the start signal reached during replay,
a message is printed, to inform the user that the interesting part of the replay
will now run. The end signal will also print a message and then disable the
replay, so that the program will continue to run by itself.
The disabling is necessary, because the rewriting ignores the rentability of
the code outside of the part of the trace, that contains the possible bug.
If the potential bug / situation is passed without crashing the program
it would most likely get stuck, because the run was altered by the rewrite.

## Trace element

To signal the end of the rewritten trace, the following element is added.

```
X,[tPost],[exitCode]
```

where `X` identifies the element as an replay control element.

- [tPost] $\in \mathbb N$: This is the time. It is replaced by the int value of the global counter at the moment when it is supposed to be run
- [exitCode]: If enabled, the replay will check if the situation described by the exit code was triggered. The exit code can have to following values:
  - 0: No exit code
  - 20: Leak: Leaking unbuffered channel or select was unstuck
  - 21: Leak: Leaking buffered channel or select was unstuck
  - 22: Leak: Leaking Mutex was unstuck
  - 23: Leak: Leaking Cond was unstuck
  - 24: Leak: Leaking WaitGroup was unstuck
  - 30: Send on close
  - 31: Receive on close
  - 32: Close on close
  - 33: Negative WaitGroup counter
  - 34: Unlock before lock
  - 41: Cyclic Deadlock

