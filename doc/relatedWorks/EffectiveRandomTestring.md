# Effective Random Testing of Concurrent Programs

[K. Sen, “Effective random testing of concurrent programs,” in Proceedings of the 22nd
IEEE/ACM International Conference on Automated Software Engineering, ser. ASE ’07,
Atlanta, Georgia, USA: Association for Computing Machinery, 2007, pp. 323–332, isbn:
9781595938824. doi: 10.1145/1321631.1321679. [Online]. Available: https://doi.org/10.1145/1321631.1321679.](https://doi.org/10.1145/1321631.1321679)

## Summary

The paper introduces an approach to testing concurrent programs by
leveraging systematic randomness in scheduling. The challenge of
testing concurrent programs lies in their non-deterministic execution,
where the same program may exhibit different behaviors across different
runs due to variations in thread scheduling. The proposes a solution
that maintains the simplicity of random testing while guiding the
execution toward potentially buggy schedules through probabilistic
interference at key synchronization points.

The central idea of the paper is to bias random testing by introducing
controlled delays at synchronization points such as lock acquisitions,
thread creations, and message passing. Instead of purely executing a
program multiple times with random thread scheduling (which is unlikely
to expose rare concurrency issues), it identifies key
points in execution where interleaving changes could have significant
effects. By selectively injecting small delays with a certain
probability, the scheduler forces different execution orders,
increasing the likelihood of uncovering concurrency-related failures.

The proposed algorithm works in several stages.
First, the program is executed once without interference, and the scheduler passively records the execution trace, noting critical synchronization operations such as mutex locks, condition variables, and thread spawns.
From this trace, the algorithm identifies specific points in execution where changes in scheduling order could lead to different outcomes.
During subsequent executions, the scheduler randomly delays some of these synchronization operations, artificially modifying the scheduling order to explore alternative interleavings.