# Trace driven dynamic deadlock detection and reproduction

[M. Samak and M. K. Ramanathan, “Trace driven dynamic deadlock detection and repro-
duction,” in Proceedings of the 19th ACM SIGPLAN Symposium on Principles and Practice
of Parallel Programming, ser. PPoPP ’14, Orlando, Florida, USA: Association for Comput-
ing Machinery, 2014, pp. 29–42, isbn: 9781450326568. doi: 10.1145/2555243.2555262.
[Online]. Available: https://doi.org/10.1145/2555243.2555262.](https://dl.acm.org/doi/10.1145/2555243.2555262)

## Summary

The paper introduces a dynamic technique for identifying and reproducing deadlocks in multithreaded programs. The key idea is to analyze execution traces to detect potential deadlocks and then use this information to reliably reproduce them. Unlike traditional static approaches that may generate false positives, or stress testing that relies on probabilistic occurrence, this method ensures that only practically occurring deadlocks are considered and that they can be deterministically triggered for debugging.

The first step of the approach involves recording execution traces of a multithreaded program, focusing on synchronization primitives such as lock acquisitions and releases. These traces capture the exact order in which threads interact with shared resources and how they acquire and release locks.

From this recorded trace, the system constructs a lock dependency graph.
If this graph contains a cycle, it indicates a potential deadlock.

Once a deadlock is detected in the trace, the method attempts to reproduce it deterministically. This is done by forcing the execution to follow the same lock acquisition order that led to the deadlock in the original trace. The system influences the thread schedule to recreate the problematic interleaving using the following mechanisms:

The system introduces scheduling constraints to ensure that threads request locks in the same order as observed in the trace.  By inserting artificial delays or controlling when threads are allowed to proceed, the system ensures that the same wait conditions develop. The execution is guided to follow the same sequence of synchronization operations seen in the trace, increasing the likelihood of hitting the deadlock again.

