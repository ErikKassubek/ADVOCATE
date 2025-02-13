
# Statistics and Time
If the `-stats` flag is set when running the `toolMain`, `toolTest` or `fuzzing` modes, multiple statistic files will be created.

If the `-time` flag is set, a file with the runtimes will be created.


> [!WARNING]
> Currently only working correctly for `toolMain`, `toolTest`. For `fuzzing` it will only contain the data of the last fuzzing run. Will be fixed soon.

## Statistics
To create the statistic files, set the `-stats` flag.\
This will create three files

- `statsTrace_progName.csv`: general statistics about the trace
- `statsAnalysis_progName.csv`: general statistics about the analysis results
- `statsAll_progName.csv`: detailed results about the trace and analysis

### statsTrace
This file contains statistics about the program traces. It contains one
line with the column names and then one line for each tests, or just one line if it is run on a main file.\
The columns are

- `TestName`: Name of the test, "Main" if main function
- `NumberOfEvents`: Total number of all events in the trace,
- `NumberOfGoroutines`: Number of go routines
- `NumberOfAtomicEvents`: Total number of atomic events,
- `NumberOfChannelEvents`: Total number of channel events,
- `NumberOfSelectEvents`: Total number of select events,
- `NumberOfMutexEvents`: Total number of mutex events
- `NumberOfWaitgroupEvents`: total number of wait group events,
- `NumberOfCondVariablesEvents`: Total number of events on conditional vars
- `NumberOfOnceOperations`: Total number of once operations

### statsAnalysis
This file contains general statistics about the program analysis. It contains one
line with the column names and then one line for each tests, or just one line if it is run on a main file.\
The columns are

- `TestName`: Name of the test, "Main" for main func
- `NumberActualBug`: Total number of actually occurring bugs
- `NumberOfLeaks`: Total number of detected leaks
- `NumberOfLeaksWithRewrite`: Total number of leaks, where a rewrite was possible
- `NumberOfLeaksResolvedViaReplay`: Total number of leaks that could be resolved with replay
- `NumberOfPanics`: Total number of possible panic bugs
- `NumberOfPanicsVerifiedViaReplay`: Total number of possible panic bugs confirmed with replay
- `NumberOfUnexpectedPanicsInReplay`: Number of unexpected bugs in replay


### statsTrace
This file contains full statistics about the program trace and analysis. It contains one
line with the column names and then one line for each tests, or just one line if it is run on a main file.\

The columns can be categories into the following sets:

- TestName: Name of the test, "Main" for main function
- Trace info
  - Total number of events
  - Number of routines (empty/not empty)
  - Number of elements per type (e.g. channel, mutex)
  - Number of operations per type
- Analysis info
  - number detected for each bug type
  - number replay created for each bug type
  - number replay successful for each bug type
  - number of unexpected panics in replay for each bug type


The full list of columns is as follows:

- `TestName`
- `NumberOfEvents`
- `NumberOfGoroutines`
- `NumberOfNotEmptyGoroutines`
- `NumberOfSpawnEvents`
- `NumberOfRoutineEndEvents`
- `NumberOfAtomics`
- `NumberOfAtomicEvents`
- `NumberOfChannels`
- `NumberOfBufferedChannels`
- `NumberOfUnbufferedChannels`
- `NumberOfChannelEvents`
- `NumberOfBufferedChannelEvents`
- `NumberOfUnbufferedChannelEvents`
- `NumberOfSelectEvents`
- `NumberOfSelectCases`
- `NumberOfSelectNonDefaultEvents`
- `NumberOfSelectDefaultEvents`
- `NumberOfMutex`
- `NumberOfMutexEvents`
- `NumberOfWaitgroup`
- `NumberOfWaitgroupEvent`
- `NumberOfCondVariables`
- `NumberOfCondVariablesEvents`
- `NumberOfOnce`
- `NumberOfOnceOperations`
- `NumberOfDetectedA00`
- `NumberOfDetectedA01`
- `NumberOfDetectedA02`
- `NumberOfDetectedA03`
- `NumberOfDetectedA04`
- `NumberOfDetectedA05`
- `NumberOfDetectedA06`
- `NumberOfDetectedA07`
- `NumberOfDetectedA08`
- `NumberOfDetectedP01`
- `NumberOfDetectedP02`
- `NumberOfDetectedP03`
- `NumberOfDetectedP04`
- `NumberOfDetectedP05`
- `NumberOfDetectedL00`
- `NumberOfDetectedL01`
- `NumberOfDetectedL02`
- `NumberOfDetectedL03`
- `NumberOfDetectedL04`
- `NumberOfDetectedL05`
- `NumberOfDetectedL06`
- `NumberOfDetectedL07`
- `NumberOfDetectedL08`
- `NumberOfDetectedL09`
- `NumberOfDetectedL10`
- `NumberOfReplayWrittenA00`
- `NumberOfReplayWrittenA01`
- `NumberOfReplayWrittenA02`
- `NumberOfReplayWrittenA03`
- `NumberOfReplayWrittenA04`
- `NumberOfReplayWrittenA05`
- `NumberOfReplayWrittenA06`
- `NumberOfReplayWrittenA07`
- `NumberOfReplayWrittenA08`
- `NumberOfReplayWrittenP01`
- `NumberOfReplayWrittenP02`
- `NumberOfReplayWrittenP03`
- `NumberOfReplayWrittenP04`
- `NumberOfReplayWrittenP05`
- `NumberOfReplayWrittenL00`
- `NumberOfReplayWrittenL01`
- `NumberOfReplayWrittenL02`
- `NumberOfReplayWrittenL03`
- `NumberOfReplayWrittenL04`
- `NumberOfReplayWrittenL05`
- `NumberOfReplayWrittenL06`
- `NumberOfReplayWrittenL07`
- `NumberOfReplayWrittenL08`
- `NumberOfReplayWrittenL09`
- `NumberOfReplayWrittenL10`
- `NumberOfReplaySuccessfulA00`
- `NumberOfReplaySuccessfulA01`
- `NumberOfReplaySuccessfulA02`
- `NumberOfReplaySuccessfulA03`
- `NumberOfReplaySuccessfulA04`
- `NumberOfReplaySuccessfulA05`
- `NumberOfReplaySuccessfulA06`
- `NumberOfReplaySuccessfulA07`
- `NumberOfReplaySuccessfulA08`
- `NumberOfReplaySuccessfulP01`
- `NumberOfReplaySuccessfulP02`
- `NumberOfReplaySuccessfulP03`
- `NumberOfReplaySuccessfulP04`
- `NumberOfReplaySuccessfulP05`
- `NumberOfReplaySuccessfulL00`
- `NumberOfReplaySuccessfulL01`
- `NumberOfReplaySuccessfulL02`
- `NumberOfReplaySuccessfulL03`
- `NumberOfReplaySuccessfulL04`
- `NumberOfReplaySuccessfulL05`
- `NumberOfReplaySuccessfulL06`
- `NumberOfReplaySuccessfulL07`
- `NumberOfReplaySuccessfulL08`
- `NumberOfReplaySuccessfulL09`
- `NumberOfReplaySuccessfulL10`
- `NumberOfUnexpectedPanicA00`
- `NumberOfUnexpectedPanicA01`
- `NumberOfUnexpectedPanicA02`
- `NumberOfUnexpectedPanicA03`
- `NumberOfUnexpectedPanicA04`
- `NumberOfUnexpectedPanicA05`
- `NumberOfUnexpectedPanicA06`
- `NumberOfUnexpectedPanicA07`
- `NumberOfUnexpectedPanicA08`
- `NumberOfUnexpectedPanicP01`
- `NumberOfUnexpectedPanicP02`
- `NumberOfUnexpectedPanicP03`
- `NumberOfUnexpectedPanicP04`
- `NumberOfUnexpectedPanicP05`
- `NumberOfUnexpectedPanicL00`
- `NumberOfUnexpectedPanicL01`
- `NumberOfUnexpectedPanicL02`
- `NumberOfUnexpectedPanicL03`
- `NumberOfUnexpectedPanicL04`
- `NumberOfUnexpectedPanicL05`
- `NumberOfUnexpectedPanicL06`
- `NumberOfUnexpectedPanicL07`
- `NumberOfUnexpectedPanicL08`
- `NumberOfUnexpectedPanicL09`
- `NumberOfUnexpectedPanicL10`

## Times
To create the time file, set the `-time` flag.\
This will create a `times_nameOfProg.csv` file in the `advocateResult` folder.
The first line of the files contains the column names. This is followed by a line containing the times for each test. If it is run on a main function, it this only contains one line.\
The columns are as follows (all times in seconds)`:

- `TestName`: Name of the test, "Main" for main function
- `ExecTime`: When running the toolchain with `-time`, the test/prog is run once without any recording or modification to get a base time
- `ExecTimeWithTracing`: Time for the trace recording
- `AnalyzerTime`: Total time for the analyzer, including reading the trace, etc.
- `AnalysisTime`: Time for only the pure trace analysis
- `HBAnalysisTime`: Time used to calculate the HB vector clocks
- `TimeToIdentifyLeaksPlusFindingPoentialPartners`: Analysis time for leaks
- `TimeToIdentifyPanicBugs`: Analysis time for panics
- `ReplayTime`: Total time for replay
- `NumberReplay`: Number of performed replays