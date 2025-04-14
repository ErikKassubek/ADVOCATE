# Trace Aware Random Testing for Distributed Systems

[B. K. Ozkan, R. Majumdar, and S. Oraee, “Trace aware random testing for distributed
systems,” Proc. ACM Program. Lang., vol. 3, no. OOPSLA, Oct. 2019. doi: 10 . 1145 /
3360606. [Online]. Available: https://doi.org/10.1145/3360606.](https://doi.org/10.1145/3360606)

## Summary

The central idea of the trace-aware random testing approach is to incorporate information about the system’s trace into the testing process.
By focusing on the execution traces that represent the system’s behavior during prior runs, the testing approach is able to make more informed decisions about which scenarios to test next.

The approach is constructed as follows:

The first step is to record the execution
traces of the system during a normal execution. These traces represent
the sequence of events, message exchanges, and state changes that occur across various nodes in the distributed system. The recorded trac captures the system's behavior, including interactions between
components, message deliveries, and resource accesses, providing
valuable context for subsequent tests.

Once the execution trace is recorded, the system uses it to generate new test scenarios.
Rather than generating random test inputs without regard to the
system's state or past interactions, the approach generates new inputs
that are contextually relevant to the recorded trace. This can involve
selecting certain states or interactions from the trace and exploring
variations of these executions.

For example, if the trace indicates a specific sequence of message
exchanges between two components, the test generator might focus on
creating tests that perturb this sequence or explore edge cases around
it. By considering the history of interactions, the testing process is
guided towards testing concurrency and interaction patterns that are
more likely to expose bugs.

The testing approach executes these generated test cases, checking for
failures or deviations from expected behavior. By focusing on execution
traces from real runs, the generated tests are more likely to hit
critical paths or scenarios where concurrency bugs such as race
conditions, deadlocks, or message ordering problems are likely to
arise. The approach increases the probability of identifying bugs that
are related to specific execution patterns.

By altering aspects of the recorded execution trace (such as
introducing message delays, reordering messages, or simulating partial
failures), the approach can simulate edge cases that might not
naturally occur during normal execution. This helps expose subtle bugs
that are related to specific timing conditions or failure modes in the
distributed system.