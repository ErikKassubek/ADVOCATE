# Evaluation Tools

This folder contains the two programs `createOverview` and `run`, to run
the programs and create an overview markdown file.

## Create Overview
This program is used to create an overview markdown file containing information
about the program, the trace, the runtimes and the analysis results.

An example for resulting overview file can be found in `exampleOutput.md`.

Create overview can be run with
```
Usage of ./overview:
  -d string
    	Path to a file with the time durations
  -n string
    	Name of the program that was analyzed
  -p string
    	Path to the program that was analyzed
  -r string
    	Path to the readable result file
  -s string
    	Path to the position where the stats file should be created
  -t string
    	Path to the trace folder
  -c boolean
      If set to true csv files are created instead of the md files
```
This means, the program run and analysis must be run before the
overview is created. This can be done by hand or using the `run`
program.

To create this we need the following files:

### -d time file

- This file contains the runtime of the different program runs.
These are 3 or 4 numbers, separated by comma. The times are
  - Time to run without advocate
  - Time to run with advocate recording
  - Time to run replay (not needed)
  - Time to run analysis
- The times file is not required

### -n name file
- The name of the program
### -p program path
- The path to the root of the program that is currently analyzed
### -r readable result
- When running the analyzer, it will create two result files called `results_readable.md` and `results_machine.md`. The r as the path to the `results_readable.md` file.
### -s result path
- Path to where the overview file should be created.
### -t path to the trace folder created by ADVOCATE

A possible program call would therefore be
```sh
./overview -d ../results/constructed/times.log -n constructed1 -p ../../examples/constructed/ -r ../results/constructed1/results_readable.log -s . -t ../../examples/constructed/trace/
```
This file will expect a times file at `../results/constructed/times.log`. The program is called `constructed1`. The program root is at `../../examples/constructed`. The result file is at `../results/constructed1/results_readable.log`. The overview file will be created at `.`. The trace of the program is stored at `../../examples/constructed/trace/`.

## Run
The `run` program is used to automatically run a program and create the
corresponding overview files.
To run the program, the following executables need to be compiled:
- ./basename_original: The original program without ADVOCATE
- ./basename_advocate: The program with ADVOCATE recording enabled
- ./basename_replay: The program with ADVOCATE replay enabled. If replay is disabled, this is not needed.

The program has the following parameters
```sh
Usage of ./run:
  -a string
        Path to advocate
  -c	Run constructed programs
  -g	Run go benchmarks
  -m	Run medium programs
  -r	Disable replay
```
There are three sets of already predefined programs. The programs
can be found in the `example` folder.
- With `-c` we can run a set of constructed programs.
- With `-g` we can run some selected `GOBench` programs.
- With `-m` we can run a collection of actual programs.

For now it is not possible to add an individual program from
the command line. If you want to add another program,
add it in the program code in one of the functions `addConstructed`,
`addGoBench` or `addMediumPrograms`. To add the program,
you have to append a list of string to the programs list.
The list of string contains (in this order)
- The name of the programs
- The path to the executables of the program that is analyzed
- The path to the root of the program that is analyzed
- The basename of the executables. The executables must have the
names `[basename]_original`, `[basename]_advocate` and `[basename]_replay` (only if `-r` is not set).
- If needed a list of the command line arguments needed for the analyzed program.

With `-a` we can set the path to the `ADVOCATE` folder (root folder of the ADVOCATE project). If it is not set, it defaults to `~/Uni/HiWi/ADVOCATE`. If it is set, set is as absolute path starting from
home. Meaning if the advocate folder is located at
`~/Uni/ADVOCATE`, set it as `-a ~/Uni/ADVOCATE`. Using relative paths will result in a failure.

With `-r` we can disable replay.

Running `run` will create a folder `results_i`, where `i` is a number.
Then for each program it will run the following steps:

- Run the base program (`basename_original`)
- Run the program with ADVOCATE recording (`basename_advocate`)
- If not disabled, run the program replay with trace from the previous step (`basename_replay`)
- Run the analysis
- Create the overview file

At the end, it will create a folder for each program in `results_i`, containing the result file.

## Statistics
Given statistic files from multiple program runs, create a latex table containing
the most relevant information.

The following tags can be set
- -f: path to a folder containing all stat file