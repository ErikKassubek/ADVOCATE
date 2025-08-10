# GoCR

GoCR is an analysis tool for concurrent Go programs.
It tries to detects blocking concurrency bugs in Go.


## Usage

> [!WARNING]
> This program currently only runs / is tested under Linux

> [!IMPORTANT]
> GoCR is implemented for go version 1.24.
> Make sure, that the program does not choose another version/toolchain and is compatible with go 1.24.
> The output `package GoCR is not in std ` or similar indicates a problem with the used version.


## Docker

We provide a docker file to create the environment.

To build the docker file, run

```shell
docker build -t gocr .
```

To run the analysis on a program, you can call the following:

```shell
docker run --rm -it \
  -v <pathToProg>:/prog \
  gocr -path /prog [args]
```

e.g.

```shell
docker run --rm -it \
  -v /home/erik/testFolder:/prog \
  gocr -path /prog -exec TestLoadConcurrent -mode GoCR
```

For the args, see [usage](#usage-1).
Note that the -path argument has already been set and does not need to be set again.


## Local

If you do not want to use the Docker container, you can also
build al the required parts yourself.


Before GoCR can be used, it must first be build.

There are two parts that need to be build.

### Runtime

To run the recording and replay for Go, a modified version of the Go runtime
has been provided. It can be found in the [go-path](../go-patch/) folder.

Before it can be used, it needs to be build. To do this, move into
[go-path/src](../go-patch/src/) directory and run the

```shell
./src/make.bash
```

script. This will create a go executable in the `bin` directory.

### GoCR

Additionally, the goCR program needs to be build. This is a standard Go
program. To build it, move into the [goCR](../goCR/) directory
and build it with the standard

```shell
go build
```

command. This will create an `gocr` executable, which will be used to
run the analysis


## Usage

To run the analysis, run the goPC program:

```shell
./goPC [args]
```

or the corresponding docker command

```shell
docker run --rm -it \
  -v <pathToProg>:/prog \
  gocr -path /prog [args]
```

The following args are required:

- \-mode [mode]
  - `GoCR`: Run our analysis version
  - `GoPie`: Run an analysis version based on GoPie [^1]
  - `GFuzz`: Run an analysis version based on GFuzz [^2]
- \-path [path]
  - path to the program to be analyzed

To run a single test, add

```shell
-exec [testName]
```

Otherwise all tests will be executed.



For additional flags, call

```shell
./goCR -help
```



[^1]: Zongze Jiang, Ming Wen, Yixin Yang, Chao Peng, Ping Yang, and Hai Jin. 2023. Effective Concurrency Testing for Go via Directional Primitive-Constrained Interleaving Exploration. In 2023 38th IEEE/ACM International Conference on
Automated Software Engineering (ASE). 1364–1376. https://doi.org/10.1109/ASE56229.2023.00086
[^2]: Ziheng Liu, Shihao Xia, Yu Liang, Linhai Song, and Hong Hu. 2022. Who goes first? detecting go concurrency bugs via message reordering. In Proceedings of the 27th ACM International Conference on Architectural Support for Programming Languages and Operating Systems (Lausanne, Switzerland) (ASPLOS ’22). Association for Computing Machinery, New York, NY, USA, 888–902. https://doi.org/10.1145/3503222.3507753