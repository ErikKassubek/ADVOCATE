# GoBench

GoBench [^1] is a benchmark for concurrency bugs in Go. It contains
multiple concurrency bugs from different real world programs.

To compare our goPie implementation with the original implementation,
we apply it to those example programs and compare the number of found bugs with
the number given in the [GFuzz paper](../../doc_proj/relatedWorks/PaperAndTools/Fuzzing/GoPie.md).

With this, we get the following results for the number of found bugs:

| GoPie Paper | GoPie Tests | GoPie Bugs | GoPie+ Tests | GoPie+ Bugs |
| --- | --- | --- | --- | --- |
| 66 | 46 | 64 | 48 | 69 |


From the table given in the GoPie paper it is not directly clear if it
shows the number of benchmark programs, where bugs were found or the
total number of bugs. Since the number given is very close to the number
of bugs measured by our implementation, we assume that it gives the total
number of found bugs.

The number of bugs detected by our direct implementation of GoPie is pretty
close to the number given in the paper. Since finding bugs always depends
on how long fuzzing is run (not really clear from the paper) and also
some lock in choosing mutations, we can say, that the numbers indicate, that
our GoPie implementation can replicate the results of the original GoPie.

The number of bugs found with GoPie+ is slightly larger than both the
values given in the GoPie paper, as well as the number of our GoPie
implementation. This may indicate, that the improvements have a positive
effect, but it cannot be ruled out, that this is also an effect of luck.

[^1]: T. Yuan, G. Li, J. Lu, C. Liu, L. Li, and J. Xue, “Gobench: A benchmark suite of real-world go concurrency bugs,” in IEEE/ACM International Symposium on Code Generation and Optimization, CGO 2021, Seoul, South Korea, February 27 - March 3, 2021, J. W. Lee, M. L. Soffa, and A. Zaks, Eds. IEEE, 2021, pp. 187–199. [Online]. Available: https://doi.org/10.1109/CGO51591.2021.9370317