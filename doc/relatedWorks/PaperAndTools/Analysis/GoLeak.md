# GoLeak

[G. -V. Saioc, D. Shirchenko and M. Chabbi, "Unveiling and Vanquishing Goroutine Leaks in Enterprise Microservices: A Dynamic Analysis Approach," 2024 IEEE/ACM International Symposium on Code Generation and Optimization (CGO), Edinburgh, United Kingdom, 2024, pp. 411-422, doi: 10.1109/CGO57630.2024.10444835.](https://ieeexplore.ieee.org/document/10444835)

## Summary

GoLeak is a diagnostic tool created to detect goroutine leaks in Go programs, particularly during the execution of tests. Internally, it works by analyzing the state of all currently running goroutines and comparing them against a reference set to identify any that are unexpectedly still active after a test concludes. Its core mechanism relies on introspection of the Go runtime via the runtime and runtime/pprof packages, which allow it to capture stack traces and metadata for all live goroutines.

When GoLeak performs its analysis, it starts by collecting a baseline snapshot of all goroutines that are known or expected to be running before a test begins or before any potentially leaking logic is executed. This is done by capturing a profile of goroutines using runtime.Stack or similar facilities, generating a detailed view of each goroutine's call stack, origin, and state.

After the test finishes, GoLeak captures a second snapshot of all active goroutines. It then performs a diff between the initial and final states to determine which goroutines were newly introduced during the test and are still running. The crucial part of this comparison is matching goroutines not just by their presence, but by analyzing their stack traces to understand their behavior and context. This comparison involves a normalization process where common non-leaking goroutines (such as those spawned by the Go runtime or long-lived background processes) are filtered out through configurable ignore rules.

The analysis focuses heavily on the top function in each goroutine’s stack trace, using it as a key identifier to classify the origin and intent of the goroutine. If a goroutine appears in the final snapshot but has no corresponding origin in the baseline and isn’t covered by ignore filters, it is flagged as a potential leak. GoLeak then reports the full stack trace of such goroutines, enabling developers to pinpoint the exact code responsible for not terminating properly.

This approach allows GoLeak to uncover subtle concurrency issues that may not cause failures during the test execution itself but indicate a larger problem in resource or lifecycle management within the program. By leveraging Go’s runtime introspection capabilities and emphasizing reproducibility in its comparisons, GoLeak serves as a powerful low-level analysis tool focused specifically on goroutine lifecycle integrity.