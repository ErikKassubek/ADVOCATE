# AdvocateGo
> [!NOTE]
> This program is a work in progress and may result in incorrect or incomplete results.

> [!WARNING]
> This program currently only runs / is tested under Linux

> [!IMPORTANT]
> ADVOCATE is implemented for go version 1.24.
> Make sure, that the program does not choose another version/toolchain and is compatible with go 1.24.
> The output `package advocate is not in std ` or similar indicates a problem with the used version.
> AdvocateGo currently does not work for programs requiring go 1.23

## What is AdvocateGo
AdvocateGo is an analysis tool for Go programs.
It detects concurrency bugs and gives  diagnostic insight.
This is achieved through `happens-before-relation` and `vector-clocks`

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
A detailed description of the inner workings can be found in the [doc](doc) folder (currently in the process of being rewritten and therefore not complete)


## Usage

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


### Analysis

The complete analysis is done with the [analyzer](analyzer).

ADVOCATE has two different modes.

#### Mode: analysis
The analysis mode is the main mode to analyzer tests or the main function of a
program.

For the (specified) tests or the main function, it will run the program and
record the trace, analyze the trace and, if something was found,
create a trace that should trigger the found bug and replay this trace,
to confirm the bug.

It can be run with
```shell
./analyzer analysis [args]
```
to run tests, or with
```shell
./analyzer analysis -main [args]
```
to analyze the main function.

The following arg is required:

- `-path [path]`: For tests, the path to the root of the project folder containing all the tests. For main, the path to the main file. Note: the program to analyzer cannot be inside the ADVOCATE folder

If the main function is analyzed, the following arg is also required:

- `-exec [path]`: Name of the executable of the program when building with `go build` (For programs that cannot simply be build with `go build` only the tests can be analyzed).

If tests are analyzed this can be used to specify a single test to be analyzed:

- `-exec [testName]`

If not set, all tests will be analyzed.

To get additional information, the following tags can also be set:

- `-recordTime`: measure the runtime for the different phases and create a time file
- `-stats`: create multiple statistic files as described [here](doc/statistics.md)\
- `-notExec`: Find operations, that have never been executed

If one of these are set, the `-prog [name]` tag must be set to indicate the name of the program.

There are additional tags. To get them, run `./analyzer analysis -h`.

While running, the analyzer will create a `advocateResult` folder. In it, it will create on
folder for each of the analyzed tests. In this folder it will create a file
for the output of the program runs, as well as two files showing an
overview over all detected bugs. Additionally, it will create a bug folder.
this folder contains one file for each of the found bugs, detailing the
type and position of the bug and information about the replay (if performed).

The created statistic and time files can also be found in the `advocateResult` folder

#### Mode: fuzzing
To run the fuzzing as described [here](doc/fuzzing.md), the following commands can be used:

```shell
./analyzer fuzzing [args]
./analyzer fuzzing -main [args]
```

The required and additional tags as well as the output files are the same as for the tool mode.

### Warning
It is the users responsibility of the user to make sure, that the input to
the program, including e.g. API calls are equal for the recording and the
tracing. Otherwise the replay is likely to get stuck.

Do not change the program code between trace recording and replay. The identification of the operations is based on the file names and lines, where the operations occur. If they get changed, the program will most likely block without terminating. If you need to change the program, you must either rerun the trace recording or change the effected trace elements in the recorded trace.
This also includes the adding of the replay header. Make sure, that it is already in the program (but commented out), when you run the recording.
