# A Trace Simplification Technique for Effective Debugging of Concurrent Programs

[N. Jalbert and K. Sen, “A trace simplification technique for effective debugging of concur-
rent programs,” Nov. 2010, pp. 57–66. doi: 10.1145/1882291.1882302](https://people.eecs.berkeley.edu/~ksen/papers/thrille.pdf)

## Summary

The paper introduces a method for reducing execution traces of concurrent programs while preserving the failure-inducing behavior. The approach starts by recording an execution trace, capturing events such as thread scheduling, synchronization operations, and shared memory accesses. This trace serves as the input for the simplification process.

The simplification is performed using an iterative reduction technique inspired by delta debugging. The system systematically removes events from the trace and re-executes the program to check if the failure is still reproducible. If an event can be removed without affecting the failure, it is discarded. If removing an event prevents the failure, it is considered essential and retained. The process continues until only the minimal subset of events necessary to reproduce the failure remains.

To ensure correctness, the approach maintains dependencies between events. Synchronization operations and ordering constraints are preserved so that the reduced trace remains a valid execution of the program. The final output is a simplified trace that retains only the critical interleavings responsible for the failure, significantly reducing the complexity of debugging. The technique is fully automated and can be applied to large-scale concurrent programs, making it easier to analyze and reproduce concurrency bugs.