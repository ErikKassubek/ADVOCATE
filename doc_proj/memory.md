# Memory Supervisor

For large traces, large amounts of data may need to be stored in RAM,
especially during the analysis. In the worst case on computers with only
limited memory, we may have more data
to store than space is available. This may fill the RAM ans Swap until they
are completely full, crashing the computer. To prevent this, a memory supervisor
has been [implemented](../advocate/memory/memory.go#L29). It runs in the
the background and continually monitors the available memory. If the memory
falls below a certain value, it will cancel the analysis of the current
program or tests. If multiple tests are analyzed, it will continue with the
next test.

The supervisor can be disabled by setting the `-noMemorySupervisor` flag.