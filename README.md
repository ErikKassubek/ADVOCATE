# AdvocateGo

## What is AdvocateGo

AdvocateGo is an analysis tool for concurrent Go programs.
It detects concurrency bugs and gives diagnostic insight.

Furthermore it is also able to produce traces which can be fed back into the program in order to experience the predicted bug.

AdvocateGo tries to detect the following situations:

- A00: Unknown panic
- A01: Send on closed channel
- A02: Receive on closed channel
- A03: Close on closed channel
- A04: Close on nil channel
- A05: Negative wait group
- A06: Unlock of not locked mutex
- A07: Concurrent recv
- A08: Select case without partner
- P01: Possible send on closed channel
- P02: Possible receive on closed channel
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
- R01: Unknown panic in recording
- R02: Timeout in recording

## Documentation

A detailed description of the inner workings can be found in the [doc](doc) folder (currently in the process of being rewritten and therefore not complete).

## Usage

> [!WARNING]
> This program currently only runs / is tested under Linux

> [!IMPORTANT]
> ADVOCATE is implemented for go version 1.24.
> Make sure, that the program does not choose another version/toolchain and is compatible with go 1.24.
> The output `package advocate is not in std ` or similar indicates a problem with the used version.

### Preparation

Before Advocate can be used, it must first be build.

First you need to build the [analyzer](analyzer).
It can be build using the standard

```shell
go build
```

command.


Additionally, the modified go runtime must be build. The runtime can be found in [go-patch](go-patch).
To build it run the

```shell
./src/make.bash
```

or

```shell
./src/make.bat
```

script. This will create a go executable in the `bin` directory.


### Analysis and Toolchain

The complete analysis and toolchain is integrated into the [analyzer](analyzer).

ADVOCATE has different modes.

#### Mode: analysis

The analysis mode is the main mode to analyzer tests.

For the (specified) tests or the main function, it will run the program and
record the trace, analyze the trace and, if something was found,
create a trace that should trigger the found bug and replay this trace,
to confirm the bug.

It can be run with

```shell
./analyzer analysis [args]
```

The following arg is required:

- `-path [path]`: For tests, the path to the root of the project folder containing all the tests. For main, the path to the main file. Note: the program to analyzer cannot be inside the ADVOCATE folder

If the main function should be analyzed instead of the unit tests, the following arg must be set (we recommend applying the analysis to the tests and not the main function):

- `-main`

For the analysis of the main function to work, the program must be buildable with `go build`. For programs that cannot simply be build with `go build` only the tests can be analyzed.

If `-main` is set and the `go.mod` file is not in the same directory as the `main.go` file, the following argument is required

- `-exec [path]`: Name of the executable of the program when building with `go build`.

If only a single test should be analyzed, this can be set with

- `-exec [testName]`

If not set, and `-main` is not set, all tests will be analyzed. Be aware, that this will run all tests
with this name, meaning if multiple tests have the same name, they will all be run.

To set timeouts, you can set

- `-timeoutRec [to in s]`: Timeout for the recording in seconds (Default: 10 min)
- `-timeoutRep [to in s]`: Timeout for the replay (Default: 500 * recording time)

To get additional information, the following tags can also be set:

- `-recordTime`: measure the runtime for the different phases and create a time file
- `-stats`: create multiple statistic files as described [here](doc/statistics.md)
- `-notExec`: Find operations, that have never been executed
- `-keepTrace`: Do not delete the trace after analysis

If one of these are set, the `-prog [name]` tag must be set to indicate the name of the program.

While running, the analyzer will create a `advocateResult` folder. In it, it will create on
folder for each of the analyzed tests. In this folder it will create a file
for the output of the program runs, as well as two files showing an
overview over all detected bugs. Additionally, it will create a bug folder.
this folder contains one file for each of the found bugs, detailing the
type and position of the bug and information about the replay (if performed).

The created statistic and time files can also be found in the `advocateResult` folder.

An example command would be

```shell
./analyzer analysis -path ~/pathToProg/progDir/ -prog progName
```

#### Mode: fuzzing

To run the fuzzing as described [here](doc/fuzzing.md), the following command can be used:

```shell
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

All other required and additional tags as well as the output files are the same as for the tool mode.

An example command would therefore be

```shell
./analyzer fuzzing -path ~/pathToProg/progDir/ -fuzzingMode GoPieHB -prog progName
```

#### Mode: Recording and Replay

During both the analysis and the fuzzing, [recording](./doc/recording.md)
and [replay](./doc/replay.md) mechanisms are used. With those, a program or
test run can be recorded and later be replayed. We also make those modes directly
available.

To create a trace from a program or test, you can run

```shell
./analyzer record [args]
```

The args must contain the `-path [pathToProg]` flag, pointing to the folder
containing the tests of the main file. All additional, applicable flags
mentioned above can also be used.

To replay a trace, you can run

```shell
./analyzer replay [args]
```

The following args are required

- `-path [pathToProg]`: path to the folder containing the tests of the main file
- `-trace [pathToTrace]`: path to the folder containing the trace files to be replayed

If the `-main` flag is not set, meaning some tests are replayed, and there
is more than one test in `pathToProg`, `-exec [testName]` must be set to specify
the test, the trace belongs to.

### Warning

It is the users responsibility of the user to make sure, that the input to
the program, including e.g. API calls are equal for the recording and the
tracing. Otherwise the replay is likely to get stuck.

Do not change the program code between trace recording and replay. The identification of the operations is based on the file names and lines, where the operations occur. If they get changed, the program will most likely block without terminating. If you need to change the program, you must either rerun the trace recording or change the effected trace elements in the recorded trace.
This also includes the adding of the replay header. Make sure, that it is already in the program (but commented out), when you run the recording.
