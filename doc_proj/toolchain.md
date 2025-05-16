# Toolchain

The toolchain allow the user to record, replay and analyze programs or tests.
It is implemented in [advocate](../advocate/).

The toolchain has 4 modi, record, replay, analysis and fuzzing.

## Recording

Recoding will take a program or a set of unit tests and [records](./recording.md)
there behavior. This will run the program or tests and record
the trace files. The files can be found in a created `advocateResult` folder.

## Replay

Replay allows us to replay a (modified) recording of a program. Given the
program or test, it will force the program execution to follow the
order specified in the trace.


## Fuzzing

Fuzzing will run a [fuzzing](fuzzing.md) approach on the tests. For this,
each test will be executed and analyzed multiple times. After each run and
analysis, the toolchain will determine if the recorded run was interesting.
For a detailed explanation of this see [here](fuzzing.md). If it was
interesting, it will create [mutations](fuzzing/mutations.md). Those mutation
specify preferred cases for selects and delays for some operations.
The toolchain will then run the full analysis again for each created mutations,
making sure that the specified restrictions are met. Each of those runs and
analyses can again create new mutations. The fuzzing will stop if no
new mutations are available, or if a predefined maximum of runs or runtime
is reached.