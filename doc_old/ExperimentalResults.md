
# Experimental results

The results for all tested programs can be found in the `experimentalResults` directory.

## Performance
To measure the performance of the modified go runtime, we ran the tracer on multiple real and constructed programs. The real programs for the performance measure are as follows:

- [etcd-io/bbolt](https://github.com/etcd-io/bbolt)
- [PuerkitoBio/gocrawl](https://github.com/PuerkitoBio/gocrawl)
- [htcat/htcat](https://github.com/htcat/htcat)
- [klauspost/pgzip](https://github.com/klauspost/pgzip)
- [jfcg/sorty](https://github.com/jfcg/sorty)
- programms from gobench (see closed on send)

To get an overview over the performance, we measure the runtimes for the programs with the unmodified go runtime as well as with the modified runtime. For the modified version we measure the pure runtime of the program (without creating the trace file) as well as the complete runtime.

The average performance overhead is calculated as 6.27% without the creation of the trace and 128.47% with the creation of the trace. But we must pay attention to the fact, that this number is massively influenced by outliers.
If we ignore the program with the highest (sorty) and lowest (grpc1687) overhead, we get an average runtime overhead of 0.3% without and 17.9% with the creation of the trace.

The runtime for the analyzer varies widely depending on the size of the trace and the operations contained in it (e.g., a trace with only atomic elements runs faster than a trace with mainly channel operations, because we do not need to search for communication partners). It is therefore hard to generalize the performance. For the tested programs, the runtime lay between 0.3s and 3 min.


## Send on Closed
Additionally we use multiple self constructed programs
as well as programs from [gobench](https://github.com/timmyyuan/gobench) to test the detection capability of the
analyzer. The programs from gobench are mainly

- grpc1687
- serving3068
- serving5865

Those programs contain potential send to closed channels. They have been partially modified to prevent actual send to closed channels.

For the programs mentioned `Performance`, no Send on Closed have been found. It is not known, if there are potential send on closed channel that have not been found, but it shows, that send on closed channel is not a very common problem. For the programs where why know, whether potential send on closed channels are present, the analyzer is able to find some but not all cases. For the gobench programs the analyzer recognizes 2 out of 3 situations. For the constructed programs, the analyzer judges 15 out of 20 programs correctly.

## Concurrent Receive
Constructed examples show, that the existents of concurrent receive operations
can be reliably detected, away in the cases discussed in the problems section.

## Problems
The program struggles mainly with the following situations:

- Order of critical sections
    - For performance reasons the analyzer does not reorder critical section. This means, that send on closed channels that require a different ordering of these critical section to occur are not detected.
- Program parts that are not run
    - The analyzer only relies on the recorded trace of the program. The trace only records those operations, that are actually performed. A send on closed, that would happen if the program would have run differently, e.g.
    because a select executed another case or an other once was executed can therefore not be detected.