# Benchmark

We provide a set of benchmarks that can be used evaluate the tool and a small runner program to run the tool.

## Benchmark Programs

With the runner program, the GoBench benchmark [^1] [^2] and a the unit tests of a set of 20 real-world programs can be run.

We provide the test of GoBench in the [./gobench](./gobench/) folder. The real-world programs are automatically cloned by the runner program when needed (make sure github is installed and that the prerequisites for running the programs unit tests are met).

The Real-World Programs are:

- [argo-cd](https://github.com/argoproj/argo-cd) (v3.1.0)
- [bleve](https://github.com/blevesearch/bleve) (v2.5.3)
- [bosun](https://github.com/bosun-monitor/bosun) (v0.8.0)
- [caddy](https://github.com/caddyserver/caddy) (v2.10.0)
- [dns](https://github.com/miekg/dns) (v1.1.50)
- [flannel](https://github.com/flannel-io/flannel) (v0.20.2)
- [frp](https://github.com/fatedier/frp) (v0.36.0)
- [gin](https://github.com/gin-gonic/gin) (v1.10.1)
- [fiber](https://github.com/gofiber/fiber) (v2.40.1)
- [gorums](https://github.com/relab/gorums) (v0.7.0)
- [grpc](https://github.com/grpc/grpc-go) (v1.51.0)
- [hugo](https://github.com/gohugoio/hugo) (v0.148.2)
- [kubernetes](https://github.com/kubernetes/kubernetes) (v1.25.5)
- [nsq](https://github.com/nsqio/nsq) (v1.3.0)
- [octant](https://github.com/vmware-archive/octant) (v0.25.1)
- [ollama](https://github.com/ollama/ollama) (v0.11.4)
- [pholcus](https://github.com/andeya/pholcus) (v1.3.4)
- [syncthing](https://github.com/syncthing/syncthing) (v1.22.1)
- [terraform](https://github.com/hashicorp/terraform) (v1.12.2)
- [zinx](https://github.com/aceld/zinx) (v1.2.7)

The programs have been taken from a range of papers reguarding concurrency bugs in Go.

Some of the tests in GoBench benchmark have been slighly modified.
Some of them have the form 

```go
func Test...(t *testing.T) {
	...
	go f() 
}
```

meaning a routine is started just before the main function returns. Since Go terminates all running goroutines as soon as the main routine returns, this will prevent the content of the created goroutine from beeing executed. This also prevents the fuzzing or analysis from running properly, since they rely on a recording of the program. We therefore modify programs like this slightly:

```go
func Test...(t *testing.T) {
	...
	go f() 

    time.Sleep(time.Second)
}
```

This allows the goroutin to execute its content.

## Usage

The runner program can be found in the [runner](./runner/) folder.

After building the [patch runtime](../goPatch/), the [controller program](../goCCT/) and the benchmark [runner](./runner/) (call `make` in the root directory), the runner program can be used to run the example programs.

To run the runner program, move into [./runner](./runner/) and execute

```shell
./runner [args]
```

The args can be as follows:

- `-mode [mode]` with possible values `record` (simple recording of the program), `analysis` (record one run and check if a bug occured) and `fuzzing`. Default: `fuzzing`.
- `-prog [prog]`. Set `gobench` to run gobench, the name of the real world program as listed above for the program, or `all` to run all programs one after another (may take a long time). Default: `gobench`.
- `-fuzzingMode [mode]` with possible values `gfuzz`, `gopie`, and `gfuzzpie`. Default: `gfuzzpie`. Has only an effect if `-mode fuzzing` is set.
- `-test [name]`: Name of the test to run. If not set, all tests are run. Default: all tests are run.

If no flags are set, the program with run fuzzing on all tests in gobench.

After the analysis is run, you can find the results in the `gocctResult` folder within the tested program folder.

[^1]: T. Yuan, G. Li, J. Lu, C. Liu, L. Li, und J. Xue, „GoBench: A Benchmark Suite of Real-World Go Concurrency Bugs“, in 2021 IEEE/ACM International Symposium on Code Generation and Optimization (CGO), Seoul, Korea (South): IEEE, Feb. 2021, S. 187–199. doi: 10.1109/CGO51591.2021.9370317.
[^2]: T. Yuan, timmyyuan/gobench. (6. August 2025). Go. Zugegriffen: 7. Januar 2026. [Online]. Verfügbar unter: [https://github.com/timmyyuan/gobench](https://github.com/timmyyuan/gobench)
