# Capturing high-level nondeterminism in concurrent programs for practical concurrency model agnostic record and replay

[D. Aumayr, S. Marr, S. Kaleba, E. Gonzalez Boix, and H. Mössenböck, “Capturing high-
level nondeterminism in concurrent programs for practical concurrency model agnostic
record amp; replay,” The Art, Science, and Engineering of Programming, vol. 5, no. 3, Feb.
2021, issn: 2473-7321. doi: 10.22152/programming-journal.org/2021/5/14. [Online].
Available: http://dx.doi.org/10.22152/programming-journal.org/2021/5/14.](https://arxiv.org/abs/2103.00031)

# Summary

The paper presents a record and replay system designed to capture nondeterministic behavior across multiple concurrency models in a unified way. Unlike traditional approaches that focus on a single concurrency paradigm, this framework enables deterministic replay for applications that combine different concurrency models, such as threads and locks, actors, communicating sequential processes (CSP), and software transactional memory (STM).

The system captures nondeterminism at a high level rather than recording low-level thread scheduling events. Instead of tracking every individual instruction or memory access, it records only the logical order of operations relevant to the concurrency model used in the application. This is achieved by identifying and logging model-specific nondeterministic events such as message sends and receives in actor-based systems, or lock acquisitions and releases in thread-based systems. Since these events define how concurrency unfolds in the program, capturing them ensures that replay can reproduce the same execution behavior while remaining independent of the underlying hardware and operating system.

To support different concurrency models in a unified way, the system introduces a structured trace representation that encodes the necessary concurrency events while preserving their dependencies. This allows the replay phase to faithfully reconstruct the execution by injecting recorded nondeterministic operations at the correct points. The system enforces the same ordering constraints observed during recording, ensuring that the program follows the same sequence of concurrency interactions, making it possible to deterministically reproduce failures that were originally caused by unpredictable thread interleavings.
