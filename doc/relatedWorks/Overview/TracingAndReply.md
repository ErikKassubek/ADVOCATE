# Tracing and Replay

The challenge of reproducing and debugging nondeterministic behavior in
concurrent programs has led to a rich body of work on record and replay systems across various concurrency models and programming paradigms. Existing approaches differ in their assumptions, overhead, and generality, with many focusing on specific memory models or concurrency abstractions.

The Go tracer [^1] is a built-in tool provided by the Go programming language for capturing and analyzing runtime events such as goroutine scheduling, system calls, garbage collection, and network blocking. It enables fine-grained performance profiling and concurrency visualization, making it particularly useful for diagnosing bottlenecks and understanding the behavior of concurrent Go programs.

"Capturing High-level Nondeterminism in Concurrent Programs
for Practical Concurrency Model Agnostic Record & Replay" [^2] presents a model-agnostic record and replay system capable of handling multi-paradigm concurrent programs. By uniformly tracing high-level nondeterministic events independent of the underlying concurrency model, this approach supports debugging in programs that combine different concurrency paradigms (e.g., actors, threads, STM). Its infrastructure allows language implementers to integrate new models without re-engineering the tracing mechanism, broadening applicability and reuse.

In contrast, CLAP [^3] addresses efficiency and minimal perturbation by logging only thread-local execution paths during runtime. Rather than capturing inter-thread interactions directly, CLAP reconstructs memory dependencies offline using constraint solving. This approach is notable for its applicability under relaxed memory models like TSO and PSO, its avoidance of additional synchronization, and its ability to derive simplified reproductions of failing executions with fewer context switches. The trade-off is the complexity of constraint solving, which CLAP mitigates through parallelization strategies.

JaRec [^4] targets Java programs specifically, relying on the JVM Profiler Interface (JVMPI) to intercept and record synchronization operations. Its design emphasizes portability and minimal JVM modification, making it a practical tool for replaying synchronization order in data-race-free Java programs. Furthermore, JaRec supports distributed execution, enabling debugging in resource-constrained or embedded environments.

Netzer [^5] proposed an adaptive tracing strategy for shared-memory programs. Their system records only memory accesses involved in nondeterministic interleavings—typically data races—achieving near-optimal trace sizes. This allows deterministic replay without assuming race freedom and dramatically reduces overhead compared to systems that log all shared-memory accesses.

Later, Netzer and Miller [^6] extend the adaptive tracing paradigm to message-passing systems. By tracing only the delivery order of messages involved in potential races, their approach reproduces the original execution with minimal overhead.

Trace-Driven Dynamic Deadlock Detection and Reproduction [^7] operates by recording execution traces during program runs and then replaying these traces under controlled schedules to verify the feasibility of potential deadlocks.

Partial Order Aware Concurrency Sampling (POS) [^8] introduces a novel sampling strategy POS. It employs a priority-based scheduling algorithm where each event is assigned a random priority. At each execution step, the event with the highest priority is executed, and after execution, all events that race with it are reassigned new random priorities.


[^1]: [The Go Authors, "Go tracer – Execution Tracing Tool", golang/go, [Online]. Available: https://pkg.go.dev/runtime/trace](./../PaperAndTools/TracingAndReplay/go-tracer.md)


[^2]: [D. Aumayr, S. Marr, S. Kaleba, E. Gonzalez Boix, and H. Mössenböck, “Capturing high-level nondeterminism in concurrent programs for practical concurrency model agnostic record amp; replay,” The Art, Science, and Engineering of Programming, vol. 5, no. 3, Feb.
2021, issn: 2473-7321. doi: 10.22152/programming-journal.org/2021/5/14. [Online]. Available: http://dx.doi.org/10.22152/programming-journal.org/2021/5/14.](./../PaperAndTools/TracingAndReplay/CapturingHighLevelNondetermanism.md)

[^3]: [J. Huang, C. Zhang, and J. Dolby, “Clap: Recording local executions to reproduce concurrency failures,” SIGPLAN Not., vol. 48, no. 6, pp. 141–152, Jun. 2013, issn: 0362-1340. doi:10.1145/2499370.2462167. [Online]. Available: https://doi.org/10.1145/2499370.
2462167.](./../PaperAndTools/TracingAndReplay/Clap.md)

[^4]: [Georges, M. Christiaens, M. Ronsse, and K. De Bosschere, “Jarec: A portable record/replay environment for multi-threaded java applications,” Software: Practice and Experience, vol. 34, no. 6, pp. 523–547, 2004. Available: https://onlinelibrary.wiley.com/doi/abs/10.1002/spe.579.](./../PaperAndTools/TracingAndReplay/JaRec.md)

[^5]: [R. H. B. Netzer, "Optimal Tracing and Replay for Debugging Shared-Memory Parallel Programs,"Brown University, 1993. doi: 10.5555/864477.](./../PaperAndTools/TracingAndReplay/OptimalTracingAndReplayMemory.md)

[^6]: [R. H. B. Netzer and B. P. Miller, "Optimal tracing and replay for debugging message-passing parallel programs," Journal of Supercomputing, vol. 8, pp. 371–388, 1995. doi: 10.1007/BF01901615.](./../PaperAndTools/TracingAndReplay/OptimalTracingAndReplayMessage.md)

[^7]: [M. Samak and M. K. Ramanathan, “Trace driven dynamic deadlock detection and reproduction,” in Proceedings of the 19th ACM SIGPLAN Symposium on Principles and Practice of Parallel Programming, ser. PPoPP ’14, Orlando, Florida, USA: Association for Computing Machinery, 2014, pp. 29–42, isbn: 9781450326568. doi: 10.1145/2555243.2555262.[Online]. Available: https://doi.org/10.1145/2555243.2555262.](./../PaPaperAndTools/TracingAndReplay/TraceDrivenDynamicDeadlockDetectionAndReproduction.md)

[^8]: [X. Yuan, J. Yang, and R. Gu, “Partial order aware concurrency sampling,” in Computer Aided Verification, H. Chockler and G. Weissenbacher, Eds., Cham: Springer International Publishing, 2018, pp. 317–335, isbn: 978-3-319-96142-2.](./../PaperAndTools/TracingAndReplay/PartialOrderAware.md)