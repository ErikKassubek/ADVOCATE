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
X,[tpost],[exitCode],[tPreLast]
```
where `X` identifies the element as an replay control element.\
- [tpost] $\in \mathbb N$: This is the time. It is replaced by the int value of the global counter at the moment when it is supposed to be run
- [exitCode]: If enabled, the replay will end with this exit code. The exit code can have to following values:
  - 0: The replay finished without being able to confirm the predicted bug
  - 3: The replay paniced unexpectadly
  - 10: Replay Stuck: Long wait time for finishing replay
  - 11: Replay Stuck: Long wait time for running element
  - 12: Replay Stuck: No traced operation has been executed for approx. 20s
  - 13: The program tried to execute an operation, although all elements in the trace have already been executed.
  - 20: Leak: Leaking unbuffered channel or select was unstuck
  - 21: Leak: Leaking buffered channel or select was unstuck
  - 22: Leak: Leaking Mutex was unstuck
  - 23: Leak: Leaking Cond was unstuck
  - 24: Leak: Leaking WaitGroup was unstuck
  - 30: Send on close
  - 31: Receive on close
  - 32: Negative WaitGroup counter
  - 33: Unlock before lock
- tPreLast $\in \mathbb N$ is the tPre of the last element in the replay, e.g. the tPre of the stuck element in a leak. Only used to detect if leak replay was successful

