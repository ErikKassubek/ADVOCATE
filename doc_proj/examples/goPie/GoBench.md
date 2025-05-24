# GoBench

GoBench [^1] is a benchmark for concurrency bugs in Go. It contains
multiple concurrency bugs from different real world programs.

To compare our goPie implementation with the original implementation,
we apply it to those example programs and compare the number of found bugs with
the number given in the [GFuzz paper](../../doc_proj/relatedWorks/PaperAndTools/Fuzzing/GoPie.md).

With this, we get the following results for the number of found bugs:

| GoPie Paper | GoPie, number of tests with found bugs | GoPie, number of fund bugs |  GoPie+, number of tests with found bugs | GoPie+, number of fund bugs |
| --- | --- | --- | --- | --- |
| 66 | 46 | 66 | 48 | 69 |


From the table given in the GoPie paper it is not directly clear if it
shows the number of benchmark programs, where bugs were found or the
total number of bugs. Since the number given is identical to the number
of bugs detected by our implementation, we assume that it gives the total
number of found bugs.

The number of bugs detected by our direct implementation of GoPie is equal to
the number given by the paper, showing that our implementation correctly
implements the functionality of the original GoPie.

The number of bugs found with GoPie+ is larger than both the
values given in the GoPie paper, as well as the number of our GoPie
implementation. We also find bugs in two more of the given tests, that the
original implementation. This indicates, that our improvements have a positive effect
on the ability of the fuzzer to find bugs.

[^1]: T. Yuan, G. Li, J. Lu, C. Liu, L. Li, and J. Xue, “Gobench: A benchmark suite of real-world go concurrency bugs,” in IEEE/ACM International Symposium on Code Generation and Optimization, CGO 2021, Seoul, South Korea, February 27 - March 3, 2021, J. W. Lee, M. L. Soffa, and A. Zaks, Eds. IEEE, 2021, pp. 187–199. [Online]. Available: https://doi.org/10.1109/CGO51591.2021.9370317