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
| Number of channel operations | 4 |
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
| Time for run without ADVOCATE | 0.003445 s |
| Time for run with ADVOCATE | 0.018414 s |
| Overhead of ADVOCATE | 434.513788 % |
| Replay without changes | 0.014712 s |
| Overhead of Replay | 327.053701 % s |
| Analysis | 0.013745 s |


## Results
==================== Summary ====================

-------------------- Critical -------------------
1 Potential leak without possible partner:
	channel: /home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:503@30
	partner: -