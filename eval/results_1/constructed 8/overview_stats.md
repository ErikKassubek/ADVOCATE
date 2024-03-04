# overview Stats

## Program
| Info | Value |
| - | - |
| Number of go files | 1 |
| Number of lines | 942 |
| Number of non-empty lines | 678 |


## Trace
| Info | Value |
| - | - |
| Number of routines | 17 |
| Number of spawns | 3 |
| Number of atomics | 0 |
| Number of atomic operations | 0 |
| Number of channels | 1 |
| Number of channel operations | 2 |
| Number of selects | 0 |
| Number of select cases | 0 |
| Number of select channel operations | 0 |
| Number of select default operations | 0 |
| Number of mutexes | 0 |
| Number of mutex operations | 0 |
| Number of wait groups | 0 |
| Number of wait group operations | 0 |
| Number of cond vars | 0 |
| Number of cond var operations | 0 |
| Number of once | 0| 
| Number of once operations | 0 |


## Times
| Info | Value |
| - | - |
| Time for run without ADVOCATE | 1.004250 s |
| Time for run with ADVOCATE | 1.020403 s |
| Overhead of ADVOCATE | 1.608464 % |
| Replay without changes | 1.019839 s |
| Overhead of Replay | 1.552303 % s |
| Analysis | 0.040326 s |


## Results
==================== Summary ====================

-------------------- Warning --------------------
1 Found receive on closed channel:
	close: /home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:172@27
	recv : /home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:169@28