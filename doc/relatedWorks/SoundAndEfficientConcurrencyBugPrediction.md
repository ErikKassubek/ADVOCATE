# Sound Predictive Fuzzing for Multi-threaded Programs

[Y. Cai, H. Yun, J. Wang, L. Qiao, and J. Palsberg, “Sound and efficient concurrency bug
prediction,” in Proceedings of the 29th ACM Joint Meeting on European Software Engi-
neering Conference and Symposium on the Foundations of Software Engineering, ser. ES-
EC/FSE 2021, Athens, Greece: Association for Computing Machinery, 2021, pp. 255–267,
isbn: 9781450385626. doi: 10.1145/3468264.3468549. [Online]. Available: https://doi.
org/10.1145/3468264.3468549](https://dl.acm.org/doi/10.1145/3468264.3468549)

# Summary

The paper address the complex challenge of detecting concurrency bugs in software systems. Traditional dynamic analysis techniques, such as the M2 approach, focus on determining whether two specific events can occur consecutively. However, real-world concurrency issues often involve multiple events and threads, with bugs manifesting when the order of two or more events can be interchanged, even if they are not consecutive.​

To tackle this, the authors introduce SeqCheck, a technique designed to soundly ascertain whether a sequence of events can occur in a specified order. This ordered sequence represents a potential concurrency bug, and various known forms of such bugs can be effectively encoded into event sequences, each depicting a possible occurrence of the bug. SeqCheck explicitly analyzes branch events and incorporates a set of efficient algorithms to achieve this. The authors demonstrate that SeqCheck is sound and complete on traces involving two threads.

The analysis uses the following steps:

- Trace collection: The program is executed once with a real input, and a trace of events is recorded. This trace includes all thread interleavings, synchronization events, and shared memory accesses.
- Building dependency graph: From the trace, a directed graph is built where nodes represent events and edges represent the HB relation. This graph captures constraints imposed by thread execution and synchronization. The paper uses three types of orders ($E^{acq}_X$: lock acquire, $E^{rd}_X$: shared memory read from):
  - Program order: $\prec_{PO}$. $\forall e_1, e_2 \in E_\sigma$, if $\text{tid}(e_1) = \text{tid}(e_2)$ and $e_1 \prec_\sigma e_2$, $\Rightarrow e_1 \prec_{PO} e_2$ (i.e., among thread-local events).
  - Observation order $\prec_{OO}$. Let $X = E_\sigma$, $\forall e \in E^{rd}_X: e = rd(t, x, w)$, $\Rightarrow w \prec_{OO} e$.
  - Lock order $\prec_{LO}$. Let $X = E_\sigma$, $\forall e_1, e_2 \in E^{acq}_X$, if $e_1 \propto e_2$, then $\text{match}_\sigma(e_1) \prec_{LO} e_2$ or $\text{match}_\sigma(e_2) \prec_{LO} e_1$.

- Event sequence matching: SeqCheck checks if the sequence of interest can be matched by reordering events in the trace, without violating the HB constraints. In other words, it looks for a legal reordering that preserves all synchronization and data dependencies but allows the sequence to occur as specified.
- Branch condition handling: A unique part of SeqCheck is its handling of branch events (e.g., conditionals that depend on shared memory). It carefully tracks which branches can flip if the order of memory accesses changes, to avoid false positives.
