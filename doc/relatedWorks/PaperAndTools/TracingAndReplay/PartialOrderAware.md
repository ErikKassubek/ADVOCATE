# Partial Order Aware Concurrency Sampling

[X. Yuan, J. Yang, and R. Gu, “Partial order aware concurrency sampling,” in Computer Aided Verification, H. Chockler and G. Weissenbacher, Eds., Cham: Springer International Publishing, 2018, pp. 317–335, isbn: 978-3-319-96142-2.](https://www.cs.columbia.edu/~rgu/publications/cav18-yuan.pdf)

# Summary

The paper presents an approach to efficiently detecting concurrency bugs by sampling thread interleavings in a structured and guided manner. The authors propose a partial order-aware sampling strategy that systematically execute interleavings.

The paper introduces BasicPOS, a priority-based scheduler and analyze its probability of covering a given partial order. Based on the
analysis of BasicPOS. It then shows that such a priority-based algorithm can be improved by introducing the priority reassignment, resulting in the POS algorithm.

The first step in POS is to record an execution trace of the program under a particular schedule. This trace includes: Thread interaction, memory accesses andynchronization events.


In BasicPOS, each event is then associated with a random and immutable priority, and, at each step in the next runs, the enabled event with the highest priority will be picked to execute.

For Pos a partial order graph (POG) is constructed. In this graph, each node represents an event. Edges between nodes represent constraints on the order of events. For instance, if one thread holds a lock and another thread is trying to acquire it, the graph would ensure that the unlock operation happens before the second thread can proceed with its lock acquisition.

This partial order graph defines causal relationships between events in the execution, meaning that some events must happen before others to maintain program correctness. Importantly, not all events are fully ordered; some may be independent and can occur in different orders without affecting correctness.

Once the partial order graph is constructed, POS samples new interleavings by perturbing the order of events while preserving the partial order constraints. Specifically, POS explores new thread interleavings by:

- Swapping independent events: Within the partial order, certain events may be independent of each other (i.e., they do not depend on one another’s completion). These independent events can be swapped in the execution order, leading to new interleavings that may trigger concurrency bugs.
- Preserving causality: The causal relationships defined in the partial order must be preserved. This means that events that are dependent on each other (e.g., one event must happen before another) cannot be reordered. The technique only changes the order of events that are not causally dependent, making sure that the perturbed execution is still valid.

This is then repeated.