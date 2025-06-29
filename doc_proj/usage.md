# How To Use Advocate


This file provides a detailed explanation on how to use the ADVOCATE framework.

## Preparation

Before Advocate can be used, it must first be build.

There are two elements that need to be build.

### Runtime

To run the recording and replay for Go, a modified version of the Go runtime
has been provided. It can be found in the [go-path](../go-patch/) folder.

Before it can be used, it needs to be build. To do this, move into
[go-path/src](../go-patch/src/) directory and run the

```shell
./src/make.bash
```

script. This will create a go executable in the `bin` directory.

### Advocate

Additionally, the advocate program needs to be build. This is a standard Go
program. To build it, move into the [advocate](../advocate/) directory
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

Normally, the print outputs of the executed programs are not printed to
the terminal, but printed into an `output.log` file. To show them on the terminal,
you can set the `-output` flag.

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
./advocate replay [args]
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


### Mode: fuzzing

To run the fuzzing as described [here](doc/fuzzing.md), the following command can be used:

```
./advocate fuzzing [args]
```

To use the fuzzing, you need to apply a fuzzing mode with `-fuzzingMode [mode]`.
The available modes are:

- `GFuzz`: Run the [GFuzz](doc/fuzzing/GFuzz.md) based fuzzing
- `GoPie`: Run the [GoPie](doc/fuzzing/GoPie.md#gopie) based fuzzing
- `GoPie+`: Run an improved [GoPie](doc/fuzzing/GoPie.md#gopie-1) based fuzzing

All other required and additional args as well as the output files are the same as for the analysis mode.

For fuzzing, the `-prog [progName]` flag with the name of the program can be set.

The number of fuzzing runs per test/prog can be limited by setting `-maxFuzzingRun [maxRun]` (default: 100). To disable this, set `-maxFuzzingRun -1`
Alternatively, a maximum time can be set using `-timeoutFuz [to in s]` (default 7 min). To disable this, set `-timeoutFuz -1`

An example command would therefore be

```
./advocate fuzzing -path ~/pathToProg/progDir/ -fuzzingMode GoPieHB -prog progName
```


## Additional Tags

To set timeouts, you can set

- `-timeoutRec [to in s]`: Timeout for the recording in seconds (Default: 10 min)
- `-timeoutRep [to in s]`: Timeout for the replay (Default: 500 * recording time)

To get additional information, the following tags can also be set:

- `-time`: measure the runtime for the different phases and create a time file
- `-stats`: create multiple statistic files as described [here](doc/statistics.md)
- `-notExec`: Find operations, that have never been executed

If one of these are set, the `-prog [name]` tag can be set to indicate the name of the program.

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

## Warning

It is the users responsibility of the user to make sure, that the input to
the program, including e.g. API calls are equal for the recording and the
tracing. Otherwise the replay is likely to get stuck.

Do not change the program code between trace recording and replay. The identification of the operations is based on the file names and lines, where the operations occur. If they get changed, the program will most likely block without terminating. If you need to change the program, you must either rerun the trace recording or change the effected trace elements in the recorded trace.
This also includes the adding of the replay header. Make sure, that it is already in the program (but commented out), when you run the recording.

## Settings

There are multiple constants that can be changed from the outside. This is
normally not necessary, but I you want to experiment with some of the settings,
you are free to do so.

Those constants can be set using the -settings [args] flag, where the args
must consists of the values that should be set in the form

```
-settings name1=value1,name2=value2
```

Make sure to not use spaces in this argument.

The following values can be changed:

| name  | default value | range | description |
|---|---|---|---|
| GFuzzW1 | 10 | $\mathbb{Q}$ | w1 weight for score in GFuzz as described [here](./fuzzing/GFuzz.md#determine-the-score) |
| GFuzzW2 | 10 | $\mathbb{Q}$ | w2 weight for score in GFuzz as described [here](./fuzzing/GFuzz.md#determine-the-score) |
| GFuzzW3 | 10 | $\mathbb{Q}$ | w3 weight for score in GFuzz as described [here](./fuzzing/GFuzz.md#determine-the-score) |
| GFuzzW4 | 10 | $\mathbb{Q}$ | w4 weight for score in GFuzz as described [here](./fuzzing/GFuzz.md#determine-the-score) |
| GFuzzFlipP | 0.99 | $\mathbb{Q}, 0 <= val <= 1$ |probability of at least one of the selects to flip as described [here](./fuzzing/GFuzz.md#flip-probability) |
| GFuzzFlipPMin | 0.1 | $\mathbb{Q}, 0 <= val <= 1$ |minimum probability for each individual select to get flipped as described [here](./fuzzing/GFuzz.md#flip-probability) |
| GoPieW1 | 1 | $\mathbb{Q}$ | w1 weight for score in GoPie as described [here](./fuzzing/GoPie.md#mutation) |
| GoPieW2 | 1 | $\mathbb{Q}$ | w2 weight for score in GoPie as described [here](./fuzzing/GoPie.md#mutation) |
| GoPieBound | 3 | $\mathbb{N}, val \geq 2$ | Maximum length of scheduling chain (BOUND) as described [here](./fuzzing/GoPie.md#mutation) |
| GoPieMutabound | 3 | $\mathbb{N}_{\neq 0}$ | Mutabound as described [here](./fuzzing/GoPie.md#mutation) |
| GoPieSCStart | 5 | $\mathbb{N}_{\neq 0}$ | Number of starting point for scheduling chains as described [here](./fuzzing/GoPie.md#mutation) |