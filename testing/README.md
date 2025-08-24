# Go Concurrency Fuzzing
This project serves as an educational overview of concurrency-related bugs in Go, exploring the effectiveness of the Go Analysis tool [ADVOCATE](https://github.com/ErikKassubek/ADVOCATE) in detecting elusive concurrency bugs in both simple practial examples and real-world projects.

---

## Go Concurrency Bugs
Go offers built-in support for concurrency through goroutines and channels, making it a popular choice for building scalable, efficent applications.

However, writing correct concurrent programs is challenging. Concurrency bugs such as race conditions, deadlocks, and improper synchronization can lead to unpredictable behavior, crashes, and security vulnerabilities. These bugs are hard to detect and reproduce due to their non-deterministic nature, since all alternative schedules that lead to the concurrency bug, need to be considered.

Fuzzing is an automated testing technique that involves providing invalid, unexpected, or random data as inputs to a program to find crashes, bugs, or unexpected behavior. While traditionally used to test input handling and robustness,in the context of concurrency, fuzzing can be used to explore different execution paths (interleavings), increasing the chances of uncovering concurrency bugs that may only occur under specific timing conditions.

```go
func TestDeadlock(t *testing.T) {
	var mu1, mu2 sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		mu1.Lock()
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
	}()

	go func() {
		defer wg.Done()
		mu2.Lock()
		mu1.Lock()
		mu1.Unlock()
		mu2.Unlock()
	}()

	wg.Wait()
}

```
Considering the classic circular wait example above. Each goroutine attempts to lock two shared mutexes, but in opposite order: one locks `mu1` then `mu2`, while the other locks `mu2` then `mu1`. The main function cannot terminate either, because it is blocked by `wg.Wait()`, which waits for both goroutines to finish signaled by `wg.Done()`: 

A deadlock occurs if the first goroutine locks `mu1` and, before it can acquire `mu2`, the second goroutine locks `mu2` and tries to acquire `mu1`. At this point, each goroutine is waiting on a lock held by the other, leading to a circular wait with no way forward, which is a classic deadlock.


```
Main          |        Go 1         |          Go 2
-----------------------------------------------------------
              | Lock mu1            |
              |                     | Lock mu2
              | Wait for mu2        | Wait for mu1
              |                     |
```

On the other hand, if one goroutine manages to acquire both locks before the second begins or reaches its second lock, the function completes successfully. This can happen, for instance, if the first goroutine locks mu1 and mu2 sequentially before the second has a chance to lock mu2.

```
Main          |        Go 1         |          Go 2
-----------------------------------------------------------
              | Lock mu1            |
              | Lock mu2            |
              | Unlock mu2          |
              |                     | Lock mu2
              |                     | Wait for mu1
              | Unlock mu1          |
              |                     | Lock mu1
```

This is where concurrency analysis tools like GFuzz, GoPie and as per latest development ADVOCATE come into play. More information on the three analysis tools are explained in [`/Docs/Tools.md`](Docs/Tools.md).


---

## Project Overview
The goal of this project is to evaluate and compare the fuzzing modes of the Advocate Go tool in terms of performance and efficiency in detecting concurrency-related bugs. It also aims to analyze overlaps and distinctions in bug detection across modes, identifying whether specific bugs are more efficeintly exposed by certain modes. 

For example, while GFuzz, GoPie, and GoPie+ can all detect select-related bugs, Advocate's GFuzz mode is expected to perform more efficiently in these cases. 


```bash
├── ADVOCATE            # Cloned ADVOCATE
├── Docs				# Documentation
│ ├── Metrics.md          	# Metrics extracted & used for comparison
│ ├── Scripts.md   			# Automation Scripts
│ ├── Tools.md				# Functionality of GFuzz, GoPie, Advocate    
│ └── Setup.md   			# Verify prerequisites
├── Examples				
│ ├── Examples_Simple      
│ ├── Examples_Projects      
├── Scripts               	# Automation scripts
├── run.py					# Run automated tests
├── config.yaml				# Config for run.py
└── README.md
```


The scripts used to automate fuzzing test cases and comparing the artefacts are explained in [`/Docs/Scripts.md`](Docs/Scripts.md). 
And the metrics extracted from the artefacts and used for comparison are explained in [`/Docs/Metrics.md`](Docs/Metrics.md).

---

### Examples_Simple
This directory contains simple Go programs that include common concurrency bugs, to test Advocate's detection of specific issues. 
They cover tests related to:
- [`Channel`](Examples/Examples_Simple/Channel/channel.md)
- [`Deadlock`](Examples/Examples_Simple/Deadlock/deadlock.md)
- [`Select`](Examples/Examples_Simple/SelectBlock/select.md)
- [`WaitGroups`](Examples/Examples_Simple/WaitGroups/waitgroups.md)
- [`Scenarios`](Examples/Examples_Simple/Scenarios/scenarios.md)

The specific test cases are each described in the respective document, containing the comparison regarding time and detected bugs across all modes and all tests. 

---

### Examples_Projects_Results
This directory contains cloned real-world Go projects to apply Advocate's analysis and fuzzing capabilities to bigger Go projects to uncover their potential concurrency issues in production-grade codebases.

| Project | Bug Types CSV | Total Time CSV |
|---------|----------------|----------------|
| [`caddy`](Examples/Examples_Projects_Results/caddy-master/) | [Bug_Types](Examples/Examples_Projects_Results/caddy-master/comparison_pivot_Bug_Types.csv) | [Time](Examples/Examples_Projects_Results/caddy-master/comparison_pivot_Total_Time_s.csv) |
| [`gin`](Examples/Examples_Projects_Results/gin-master/)     | [Bug_Types](Examples/Examples_Projects_Results/gin-master/comparison_pivot_Bug_Types.csv)   | [Time](Examples/Examples_Projects_Results/gin-master/comparison_pivot_Total_Time_s.csv)   |
| [`cobra`](Examples/Examples_Projects_Results/cobra-main/)   | [Bug_Types](Examples/Examples_Projects_Results/cobra-main/comparison_pivot_Bug_Types.csv)   | [Time](Examples/Examples_Projects_Results/cobra-main/comparison_pivot_Total_Time_s.csv)   |


---

### Quick Start

#### Run
To execute automated fuzzing tests, use the `run.py` script with the path to a target project directory. Fuzzing configuration is specified in `config.yaml`.

```bash
python3 ./run.py ./testing/Examples/Examples_Simple/SelectBlock/
```

You will be prompted to select the automation mode. 

```
========== Advocate Fuzzing Runner ==========
Select mode:
  1) Run ALL modes on ALL test cases
  2) Run ALL modes on ONE selected test case
  3) Run ONE selected mode on ALL test cases
  4) Run ONE selected mode on ONE selected test case
  5) Exit
Choice [1-5]: 
```
For options 1 and 2, all fuzzing modes listed in the config will be executed.
For options 3 and 4, you can select a specific fuzzing mode:

```bash
Select fuzz mode:
1) GFuzz
2) GFuzzHB
3) Flow
4) GFuzzHBFlow
5) GoPie
6) GoPie+
7) GoPieHB
#? 1 // Inputing number 1 selects GFuzz
```

For options 2 and 4, an additional prompt will let you choose which test to run, as all test cases are searched for in the provided test directory. Option 2 runs all modes on the selected test and option 4 runs the selected mode on the selected test.

**Example**: Found tests within `Examples/Examples_Simple/Channel/`
```bash
1) TestBufferedFillNoRead
2) TestBufferedDrainSlow
3) TestUnbufferedSendRecv
4) TestUnbufferedLeakNoRecv
5) TestUnbufferedRecvNoSend
#? 1 // Inputing number 1 selects the first test case
```



#### Results
In all cases a `/results` directory is created within the directory given as argument to the `run.py`.

Each test case (e.g., `TestBufferedFillNoRead`) gets its own subfolder under `/results/`. Inside that, results for each fuzzing mode (e.g., `GFuzz`, `GoPie+`) are stored in separate subdirectories named after the mode.

If you selected **option 1** (all tests, all modes) or **option 2** (one test, all modes), a `combined/` folder and a `comparison.csv` are generated inside the test's results folder. These allow comparing all modes for that test using defined metrics.

For a full explanation of the scripts, see [`/Docs/Scripts.md`](Docs/Scripts.md).


**Example:**
```
Channel/
├── results/
│ └── TestBufferedFillNoRead/
│ 	├── combined/
│ 	├── GFuzz/
│ 	├── GFuzzHB/
│ 	├── GFuzzHBFlow/
│ 	├── GoPie/
│ 	├── GoPie+/
│ 	├── GoPieHB/
│ 	└── comparison.csv
| └── TestBufferedDrainSlow/
│ 	├── combined/
│ 	├── GFuzz/
|   ├── ...
```


#### Comparing Specific Metrics
While each test's `comparison.csv` and `combined/` directory summarize all modes **for that specific test only**, you can generate **cross-test comparisons** of specific metrics using the script `compare_all_by_one.py`:

```bash
python3 ./testing/Scripts/compare/compare_all_by_one.py ./testing/Examples/Examples_Simple/SelectBlock/results/
```

This script scans for existing **comparison.csv** in each test case folder under a given `results/` directory and builds a **pivot table** comparing one metric across all tests and modes.

You will be prompted to select a metric:

```bash
Available metrics:
1) Ana_s
2) Bug_Types
3) Bugs_per_Minute
4) Confirmed_Replays
5) Leaks
6) Panics
7) Rec_s
8) Rep_s
9) Replays_Successful
10) Replays_Written
11) Runs_per_Minute
12) Total_Bugs
13) Total_Runs
14) Total_Time_s
15) Unique_Bugs
Select metric by number: # Inputing 14 creates a Pivot table for metric "Total_Time_s"
```

The result is a `comparison_pivot_[METRIC].csv` which is saved directly under the `/results/` directory you provided. Here is an [example](Examples/Examples_Simple/SelectBlock/results/comparison_pivot_Total_Time_s.csv). This file lets you easily compare the selected metric across all tests and modes in tabular form.

For a full explanation of the metrics, see [`/Docs/Metrics.md`](Docs/Metrics.md).