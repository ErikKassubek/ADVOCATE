# Fuzzing

DeadlockFuzzer [^1] introduces a two-stage approach to detect real deadlocks in Java programs. It combines lightweight dynamic analysis to identify potential deadlock scenarios and then utilizes a randomized thread scheduler to increase the likelihood of manifesting those deadlocks in practice. This method avoids false positives, a common drawback of static deadlock detection tools, by validating that the identified cycles can be exercised in real executions.

GFuzz [^2] focuses on Go programs and detects channel-related bugs. It focuses on mutating the cases executed in selects. The basic idea lies in setting a preferred case for each select. After executing it is checked, whether the run explored new operations. If so, the set of preferred cases is mutated, to get the next mutation run.

GoPie [^3] expands this idea to by directly influencing the interleaving of channel messages and mutexes to detect concurrency bugs. To prevent the path explosion problem, GoPie divides the program into smaller fragments, that can be individually mutated. It utilizes execution histories to identify new interleavings instead of relying on exhaustive exploration or random scheduling.


PREDFUZZ [^4] extends AFL++ with a predictive analysis phase, selectively invoking tools like SEQCHECK to anticipate concurrency vulnerabilities even if they are not directly observed during a test run. This hybrid of predictive and feedback-directed fuzzing offers improved coverage with soundness guarantees.

Greybox Fuzzing for Concurrency Testing [^5] takes another direction by adapting greybox fuzzing to the interleaving space. It applies biased random search guided by the “reads-from” relation to explore diverse and behaviorally distinct thread schedules. This strategy efficiently uncovers bugs compared to traditional random or exhaustive schedule exploration methods.

Randomized and Partial-Order-Based Testing. Random testing is a simple yet effective strategy to explore nondeterministic schedules. Sen [^6] highlights the limitations of naive random scheduling and proposes RAPOS, a partial order sampling algorithm that improves the distribution of explored schedules by reducing sampling bias. This technique has shown to outperform purely random testing in coverage and bug discovery.

Trace-Aware Random Testing [^7] further extends this idea by combining partial-order reduction with bug-depth-based schedule sampling. The proposed taPCT algorithm samples low-depth executions, which are empirically more likely to reveal bugs, while being trace-aware to avoid redundant or irrelevant interleavings. With theoretical guarantees and empirical validation on distributed systems like Cassandra and Zookeeper, this approach demonstrates high efficiency in uncovering deep concurrency bugs.

[^1]: [P. Joshi, C.-S. Park, K. Sen, and M. Naik, “A randomized dynamic program analysis
technique for detecting real deadlocks,” SIGPLAN Not., vol. 44, no. 6, pp. 110–120, Jun.
2009, issn: 0362-1340. doi: 10.1145/1543135.1542489. [Online]. Available: https://doi.org/10.1145/1543135.1542489](./../PaperAndTools/Fuzzing/DeadlockFuzzer.md)

[^2]: [Z. Liu, S. Xia, Y. Liang, L. Song, and H. Hu, “Who goes first? detecting go concurrency bugs via message reordering,” in Proceedings of the 27th ACM International Conference on Architectural Support for Programming Languages and Operating Systems, ser. ASPLOS ’22, Lausanne, Switzerland: Association for Computing Machinery, 2022, pp. 888–902, isbn: 9781450392051. doi: 10.1145/3503222.3507753. [Online]. Available: https://doi.org/10.1145/3503222.3507753](./../PaperAndTools/Fuzzing/GFuzz.md)

[^3]: [Z. Jiang, M. Wen, Y. Yang, C. Peng, P. Yang and H. Jin, "Effective Concurrency Testing for Go via Directional Primitive-Constrained Interleaving Exploration," 2023 38th IEEE/ACM International Conference on Automated Software Engineering (ASE), Luxembourg, Luxembourg, 2023, pp. 1364-1376, doi: 10.1109/ASE56229.2023.00086.](./../PaperAndTools/Fuzzing/GoPie.md)

[^4]: [Y. Guo et al., "Sound Predictive Fuzzing for Multi-threaded Programs," 2023 IEEE 47th Annual Computers, Software, and Applications Conference (COMPSAC), Torino, Italy, 2023, pp. 810-819, doi: 10.1109/COMPSAC57700.2023.00110.](./../PaperAndTools/Fuzzing/SoundPreditciveFuzzing.md)

[^5]: [D. Wolff, Z. Shi, G. J. Duck, U. Mathur, and A. Roychoudhury, “Greybox fuzzing for con-
currency testing,” in Proceedings of the 29th ACM International Conference on Architec-
tural Support for Programming Languages and Operating Systems, Volume 2, ser. ASPLOS
’24, La Jolla, CA, USA: Association for Computing Machinery, 2024, pp. 482–498, isbn:
9798400703850. doi: 10.1145/3620665.3640389. [Online]. Available: https://doi.org/10.1145/3620665.3640389.](./../PaperAndTools/Fuzzing/GreyboxFuzzing.md)

[^6]: [K. Sen, “Effective random testing of concurrent programs,” in Proceedings of the 22nd IEEE/ACM International Conference on Automated Software Engineering, ser. ASE ’07, Atlanta, Georgia, USA: Association for Computing Machinery, 2007, pp. 323–332, isbn: 9781595938824. doi: 10.1145/1321631.1321679. [Online]. Available: https://doi.org/10.1145/1321631.1321679.](./../PaperAndTools/Fuzzing/EffectiveRandomTestring.md)

[^7]: [B. K. Ozkan, R. Majumdar, and S. Oraee, “Trace aware random testing for distributed systems,” Proc. ACM Program. Lang., vol. 3, no. OOPSLA, Oct. 2019. doi: 10.1145/3360606. [Online]. Available: https://doi.org/10.1145/3360606.](./../PaperAndTools/Fuzzing/TraceAwareRandomTesting.md)