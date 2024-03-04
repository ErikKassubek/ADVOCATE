# overview Stats

## Program
| Info | Value |
| - | - |
| Number of go files | 9 |
| Number of lines | 1505 |
| Number of non-empty lines | 958 |


## Trace
| Info | Value |
| - | - |
| Number of routines | 32 |
| Number of spawns | 18 |
| Number of atomics | 389 |
| Number of atomic operations | 4295 |
| Number of channels | 38 |
| Number of channel operations | 52 |
| Number of selects | 54 |
| Number of select cases | 111 |
| Number of select channel operations | 59 |
| Number of select default operations | 26 |
| Number of mutexes | 58 |
| Number of mutex operations | 435 |
| Number of wait groups | 4 |
| Number of wait group operations | 19 |
| Number of cond vars | 2 |
| Number of cond var operations | 11 |
| Number of once | 24| 
| Number of once operations | 574 |


## Times
| Info | Value |
| - | - |
| Time for run without ADVOCATE | 1.256031 s |
| Time for run with ADVOCATE | 1.289844 s |
| Overhead of ADVOCATE | 2.692051 % |
| Analysis | 0.052687 s |


## Results
==================== Summary ====================

-------------------- Critical -------------------
1 Potential mixed deadlock:
	locks: 
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/log/log.go:243@6531
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/log/log.go:243@6517
	send/close-recv: 
		/home/erikkassubek/Uni/HiWi/Other/examples/gocrawl/crawler.go:295@6514
		/home/erikkassubek/Uni/HiWi/Other/examples/gocrawl/worker.go:66@6537
2 Possible negative waitgroup counter:
	done: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:655@879
	done/add: 
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:655@1616
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:655@1653
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:655@916
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/lookup.go:358@1667
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/lookup.go:358@930
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:651@1447
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:651@1450
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/lookup.go:322@1391
3 Possible negative waitgroup counter:
	done: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:655@916
	done/add: 
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:655@1616
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:655@1653
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:655@879
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/lookup.go:358@1667
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/lookup.go:358@930
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:651@1447
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/dnsclient_unix.go:651@1450
		/home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/lookup.go:322@1391
-------------------- Warning --------------------
4 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/fd_unix.go:105@955
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/fd_unix.go:118@954
5 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/transport.go:1242@1001
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/transport.go:1389@498
6 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/fd_unix.go:105@1692
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/fd_unix.go:118@1691
7 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/crypto/internal/randutil/randutil.go:28@1758
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/crypto/internal/randutil/randutil.go:31@1764
8 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/h2_bundle.go:931@5446
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/h2_bundle.go:904@5416
9 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/transport.go:1242@5451
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/transport.go:1389@1383
10 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/h2_bundle.go:9374@6002
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/h2_bundle.go:8312@5503
11 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/h2_bundle.go:9760@6206
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/h2_bundle.go:8477@5740
12 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/h2_bundle.go:8609@6280
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/go-patch/src/net/http/h2_bundle.go:9619@6492
13 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/Other/examples/gocrawl/crawler.go:295@6514
	recv : /home/erikkassubek/Uni/HiWi/Other/examples/gocrawl/worker.go:66@6537