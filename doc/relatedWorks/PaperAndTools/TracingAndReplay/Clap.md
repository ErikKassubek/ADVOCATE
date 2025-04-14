# CLAP: Recording Local Executions to Reproduce Concurrency Failures

[J. Huang, C. Zhang, and J. Dolby, “Clap: Recording local executions to reproduce concur-
rency failures,” SIGPLAN Not., vol. 48, no. 6, pp. 141–152, Jun. 2013, issn: 0362-1340. doi:
10.1145/2499370.2462167. [Online]. Available: https://doi.org/10.1145/2499370.
2462167.](https://dl.acm.org/doi/10.1145/2499370.2462167)

## Summary

The goal of CLAP is to efficiently reproduce bugs (mainly asserts) in multi-threaded programs, while mainly focusing on reducing the overhead of the trace recording.

CLAP addresses this challenge by recording trace local executions of a program including a bug and then solving a constraint problem to create
a global trace. The solving of the constraint problem has multiple steps:

- Find all the possible shared data access points (called SAP - a read, write, or synchronization) on the thread local paths that may cause non-determinism, via a static escape analysis.
- Compute the path conditions for each thread with symbolic execution. Given the program input, the path conditions are all symbolic formulae with the unknown values read by the SAPs.
- Encode all the other necessary execution constraints – i.e., the bug manifestation, the synchronization order, the memory order, and the read-write constraints – into a set of formulae in terms of the symbolic value variables and the order variables.
Use a SMT solver to solve the constraints, which computes a schedule represented by an ordering of all the SAPs, and this schedule is then used by an application-level.

<center><img src="../img/clap.png" alt="Order enforcement" width="800px" height=auto></center>

Based on the reconstructed trace, CLAP is then able to replay bugs.