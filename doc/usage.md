# How To Use Advocate

<!-- Rewrite Mode Recording -->
<!-- Rewrite Mode Replay -->
<!-- Rewrite Mode Analysis -->
<!-- Rewrite Mode Fuzzing -->

This file provides a detailed explanation on how to use the ADVOCATE framework.

## Preparation

Before Advocate can be used, it must first be build.

There are two elements that need to be build.

### Runtime

To run the recording and replay for Go, a modified version of the Go runtime
has been provided. It can be found in the [go-path](../../go-patch/) folder.

Before it can be used, it needs to be build. To do this, move into
[go-path/src](../../go-patch/src/) directory and run the

```shell
./src/make.bash
```

script. This will create a go executable in the `bin` directory.

### Advocate

Additionally, the advocate program needs to be build. This is a standard Go
program. To build it, move into the [advocate](../../advocate/) directory
and build it with the standard

```shell
go build
```

command. This will create an `advocate` executable, which will be used to
run all recordings, replays, analysis and fuzzing.

> [!NOTE]
> ADVOCATE is implemented for [go version 1.24](https://go.dev/blog/go1.24).\
> Make sure, that the correct version is installed on your system.\
> Make sure, that the executed programs and tests do not choose another version/toolchain and are compatible with go 1.24.\
> The output `package advocate is not in std ` or similar indicates a problem with the used version.


## Usage

All modes of advocates are started and controlled using the [advocate](../advocate/)
program. This program implements multiple modes:

- [Recording](#mode-recording)
- [Replay](#mode-replay)
- [Analysis](#mode-analysis)
- [Fuzzing](#mode-fuzzing)

### Mode: Recording

This mode allows us to record all relevant concurrency operations of a given
program. Running this mode will run a given program or test and create a
trace of all the elements executed in the program.

For more information about the recording and the created trace, see [here](./recording.md).

To run a recording, we use the advocate program implemented [here](../advocate/).
To record a program, run

```shell
./advocate record [args]
```

This will run the program or tests and create the traces. They are placed
into a folder called `advocateResult`.

#### Tests

If you want to execute and record the unit tests of a program, you need to
specify the following arg:

- `-path [pathToTests]`

This provides the path to the tests that should be recorded.
The path should point to the folder containing the tests.

As a default, this will record all tests
in this folder (for each test a separate trace will be created).\
If you only want to execute a specific test, you can specify this test with

- `-exec [testName]`

Be aware that this will record all tests with this name, meaning if multiple
tests share the same name, all of them will be executed.

An example command would be

```
./advocate record -path ~/program/testFolder/ -exec TestOne
```

#### Program

If you want to record the program itself, the following args need to be set

- `-main`
- `-path [pathToMainFile]`

Here, main tells the program to run the main function instead of the tests.
Path should point to the file containing the main function of this program.

Go will try to determine the executable name of the program from the `go.mod`
file. If advocate is unable to find this file, it needs to be set manually
using

- `-exec [executableName]`

Please note, that programs that cannot be directly be build using `go build`
cannot be recorded (the tests can still be recorded in this case).

A possible command would therefore be

```
./advocate record -main -path ~/program/main.go -exec progName
```

### Mode: Replay

The replay mode allows us to replay a previously recorded trace.

This can be done by calling

```
./analyzer replay [args]
```

The following args are required

- `-path [pathToProg]`: path to the folder containing the tests of the main file
- `-trace [pathToTrace]`: path to the folder containing the trace files to be replayed

If the `-main` flag is not set, meaning some tests are replayed, and there
is more than one test in `pathToProg`, `-exec [testName]` must be set to specify
the test, the trace belongs to.\
For main, `-exec [executableName]` should only be set if the go.mod file cannot
be found.

Possible command would therefore be

```
./advocate replay -path ~/program/testFolder/ -trace ~/traceFolder
./advocate -main replay -path ~/program/main.go -trace ~/traceFolder
```

Please not, that the trace folder should not be inside the `AdvocateResult` folder.
This means, if you want to replay a given recording, you first need to copy it
to a location outside the `AdvocateResult` folder.

Please note, that the replay relies on the program code not being altered
between recording and replay. Each change, even on non-concurrency elements
can cause the replay to fail.\
Additionally, all non-concurrency indeterminism, like random numbers
or unpredictable api calls can cause the replay to get stuck. For more info
see [here](./replay.md#things-that-can-go-wrong).

### Mode: analysis

The analysis mode is the main mode to analyzer tests. It will run the program
or test, [record](./recording.md) the trace, [analyze](./analysis.md) and analyze
it. It a potential bug was found and it is possible, the mode will [rewrite](./rewrite.md)
the trace in such a way, that it should contain the potential bug and
[replay](./replay.md) it, trying to trigger the bug.

For the (specified) tests or the main function, it will run the program and
record the trace, analyze the trace and, if something was found,
create a trace that should trigger the found bug and replay this trace,
to confirm the bug.

It can be run with

```
./advocate analysis [args]
```

The arguments for `-path` and if necessary `-main` are required. They are
defined the same way as in for the [recording](#mode-recording) and [replay](#mode-replay)
modes.

The default behavior is to run all analysis scenarios. You can select to run
only certain scenarios to by setting

- `-scen [scenarios]`

with the following possible scenarios:

- `s`: Send on closed channel
- `r`: Receive on closed channel
- `w`: Done before add on waitGroup
- `n`: Close of closed channel
- `b`: Concurrent receive on channel
- `l`: Leaking routine
- `u`: Unlock of unlocked mutex
- `c`: Cyclic deadlock (resource deadlocks)

To select multiple by adding them together, e.g.

```
  -scen sc
```

to run the analysis for send on closed and cyclic (resource) deadlocks.\
If `-scen` is not set, all scenarios will be searched for.

While running, the analyzer will create a `advocateResult` folder. In it, it will create on
folder for each of the analyzed tests. In this folder it will create a file
for the output of the program runs, as well as two files showing an
overview over all detected bugs. Additionally, it will create a bug folder.
This folder contains one file for each of the found bugs, detailing the
type and position of the bug and information about the replay (if performed).

An example command would be

```
./analyzer analysis -path ~/pathToProg/progDir/main.go -prog progName -main
```

to run the analysis on the main function of a program, or

```
./analyzer analysis -path ~/pathToProg/progDir/ -prog progName -scen c -exec TestOne
```

to analyze the test `TestOne` in the given path, only checking for cyclic (resource) deadlocks.

The analysis will try to find the following situations:

- A01: Send on closed channel
- A02: Receive on closed channel (only warning)
- A03: Close on closed channel
- A04: Close on nil channel
- A05: Negative wait group
- A06: Unlock of not locked mutex
- A07: Concurrent recv
- P01: Possible send on closed channel
- P02: Possible receive on closed channel (only warning)
- P03: Possible negative waitgroup counter
- L01: Leak on unbuffered channel with possible partner
- L02: Leak on unbuffered channel without possible partner
- L03: Leak on buffered channel with possible partner
- L04: Leak on buffered channel without possible partner
- L05: Leak on nil channel
- L06: Leak on select with possible partner
- L07: Leak on select without possible partner (includes nil channels)
- L08: Leak on mutex
- L09: Leak on waitgroup
- L10: Leak on cond

Some of them are only considered warnings. To ignore them, you can set `-noWarning`.

The analysis will try to rewrite and replay found potential bugs. To disable this,
you can set `-noRewrite`.\
The default behavior is to not replay bugs that have already been replayed successfully.
To still replay them, you can set `-replayAll`.

The traces can become very large. When using advocate for many tests, or multiple
times, this can lead to a large amount of data being stored in the trace files.
For this reason, advocate will delete the trace files, as soon as the analysis
has finished. To keep the traces, you can set `-keepTrace`.

The [Go Memory-Model](https://go.dev/ref/mem#chan) does not specify, that
channels behave as a FIFO queue. Since in practice they are implemented as such,
the analysis assumes that they behave this way. If you do not want to use this
assumption, you can disable it using `-ignCritSec`.

Additionally, the used happens-before modes assumes, that critical sections
(mutex) conform to a happens before relation, meaning they cannot be reordered.
This may be too strong of an assumption in some case. To ignore this relation
you can set `-ignCritSec`.

If you do not want to perform the happens-before analysis, but only check
for actually occurring panics or leaks, you can set the `-onlyActual` flag.

### Mode: fuzzing

To run the fuzzing as described [here](doc/fuzzing.md), the following command can be used:

```
./analyzer fuzzing [args]
```

To use the fuzzing, you need to apply a fuzzing mode with `-fuzzingMode [mode]`.
The available modes are:

- `GFuzz`: Run the [GFuzz](doc/fuzzing/GFuzz.md) based fuzzing
- `GFuzzHB`: Run the improved [GFuzz](doc/fuzzing/GFuzz.md#improvement-over-original-gfuzz) based fuzzing using happens-before information
- `Flow`: Run the [Flow](doc/fuzzing/Flow.md) based fuzzing
- `GFuzzHBFlow`: Run a combination of [GFuzzHB](doc/fuzzing/GFuzz.md) and the [Flow](doc/fuzzing/Flow.md) based fuzzing
- `GoPie`: Run the [GoPie](doc/fuzzing/GoPie.md#gopie) based fuzzing
- `GoPie+`: Run an improved [GoPie](doc/fuzzing/GoPie.md#gopie-1) based fuzzing
- `GoPieHB`: Run an improved [GoPie](doc/fuzzing/GoPie.md#gopiehb) based fuzzing using happens-before information

All other required and additional args as well as the output files are the same as for the analysis mode.

An example command would therefore be

```
./analyzer fuzzing -path ~/pathToProg/progDir/ -fuzzingMode GoPieHB -prog progName
```


## Additional Tags

To set timeouts, you can set

- `-timeoutRec [to in s]`: Timeout for the recording in seconds (Default: 10 min)
- `-timeoutRep [to in s]`: Timeout for the replay (Default: 500 * recording time)

To get additional information, the following tags can also be set:

- `-time`: measure the runtime for the different phases and create a time file
- `-stats`: create multiple statistic files as described [here](doc/statistics.md)
- `-notExec`: Find operations, that have never been executed

If one of these are set, the `-prog [name]` tag must be set to indicate the name of the program.

The created statistic and time files can also be found in the `advocateResult` folder.

In some situations, especially when only limited storage is available, it may
be useful to ignore atomic operations during recording and analysis. To do this,
you can set the `-ignoreAtomics`.

Insufficient memory (RAM) can cause the computer running advocate to crash
during the analysis. To stop this, a [Memory Supervisor](./memory.md) has been implemented.
Disable this supervisor, you can set the `- noMemorySupervisor` flag.

If the analysis of multiple tests was interrupted, running the toolchain
again would start from the beginning. If you want to skip all the already
finished tests, you can set `-cont`.

During the runtime, multiple status messages are shown in the terminal.
To disable them and only show found bugs, you can set `-noInfo`.

Sometimes the analysis or execution of a test may result in a panic in
advocate. Since it would be annoying to terminate the total analysis of all tests,
only because the analysis of one of the tests crashed, a catch mechanism
has been implemented, that will only terminate the analysis of the given tests
and continue with the next, without crashing the whole program. To disable this,
you can set the `-panic` flag.