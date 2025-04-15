# Optimal Tracing and Replay for Debugging Shared-Memory Parallel Programs

[R. H. B. Netzer, "Optimal Tracing and Replay for Debugging Shared-Memory Parallel Programs,"Brown University, 1993. doi: 10.5555/864477.](https://dl.acm.org/doi/10.5555/864477)

## Summary

The paper addresses the challenge of enabling deterministic replay for debugging nondeterministic behavior in explicitly parallel, shared-memory programs. Such programs can exhibit subtle concurrency bugs due to data races, where the outcome of execution depends on the unpredictable interleaving of memory accesses by multiple threads.

Traditional record and replay techniques often incur significant overhead because they conservatively trace all shared-memory operations to ensure faithful reproduction. Netzer and Miller propose a more efficient and elegant solution: an adaptive tracing strategy that records only the minimal set of memory operations required to deterministically replay an execution. The core insight is that not all memory accesses need to be logged—only those involved in race conditions that can influence execution outcomes.

Their technique dynamically detects these critical data races during program execution, without prior knowledge of program correctness. This allows the system to operate without assuming that the program is race-free, a limitation of many previous methods. When a potential race is detected, the tracing mechanism records only the information necessary to preserve the order of the conflicting accesses during replay.

The result is a significant reduction in trace size and runtime overhead. Experiments demonstrate that the system can reduce logging overhead by one to two orders of magnitude compared to approaches that trace all shared-memory references. This efficiency enables practical use even for large and long-running programs, making it feasible to apply deterministic replay to a wider range of debugging scenarios.