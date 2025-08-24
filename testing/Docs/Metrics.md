# Metrics Overview

This describes the **metrics** extracted from the Advocate results during fuzzing runs, where they originate from, and what they represent. These metrics are compiled into `comparison.csv` by `compare_one_by_all.py` and used for cross-test analysis by `compare_all_by_one.py`.

---

## 1. Unique_Bugs
The number of unique bugs detected during fuzzing. 
To indicate how many distinct bugs were found across all fuzzing runs for a given mode. 
Extracted from `statsAll_*.csv`, specifically columns beginning with `NoUniqueDetected*`.

---

## 2. Bug_Types
A semicolon-separated list of bug types that were detected. Shows the types of unique bugs, such as deadlocks, data races, or specific concurrency errors.
Extracted from `statsAll_*.csv`, column headers starting with `NoUniqueDetected`.

---

## 3. Total_Bugs
The total number of bugs (unique or repeated) encountered. 
Overall count of all bug occurrences across fuzzing runs.
Extracted from `statsAnalysis_*.csv`, second column (`TotalBugs`).

---

## 4. Panics
Number of test executions that resulted in a runtime panic.
Indicates test executions that crashed due to unexpected runtime errors.
Extracted from `statsAnalysis_*.csv`, column `Panics`.

---

## 5. Leaks
Number of concurrency-related resource leaks detected (e.g., goroutine leaks).
Helps identify tests where resources are not properly released.
Extracted from `statsAnalysis_*.csv`, column `Leaks`.

---

## 6. Confirmed_Replays
Number of confirmed bug replays that succeeded.
Ensures that detected bugs are reproducible and not false positives.
Extracted from `statsAnalysis_*.csv`, column `ConfirmedReplays`.

---

## 7. Total_Runs
Total number of fuzzing runs executed for the test.
Indicates test coverage and the amount of fuzzing effort applied.
Extracted from `statsFuzzing_*.csv`, second column (`TotalRuns`).

---

## 8. Total_Time_s
Total execution time in seconds for the entire fuzzing run.
Measures total time cost for fuzzing this test in the given mode.
Extracted from `times_total_*.csv`, second column (`TotalTime`).

---

## 9. Rec_s
Recording phase time in seconds.
Duration spent in the trace-recording phase of the fuzzing workflow.
Extracted from `times_detail_*.csv`, column `Recording`.

---

## 10. Ana_s
Analysis phase time in seconds.
Duration spent analyzing recorded execution traces for errors.
Extracted from `times_detail_*.csv`, column `Analysis`.

---

## 11. Rep_s
Replay phase time in seconds.
Extracted from `times_detail_*.csv`, column `Replay`.
Duration spent replaying identified issues to confirm correctness.

---

## 12. Bugs_per_1000_Runs
Number of bugs per 1000 fuzzing runs.
Derived metric computed as `(Total_Bugs / Total_Runs) * 1000`.
Normalizes bug discovery by test effort, enabling fair comparison across different run counts.

---

## 13. Bugs_per_Minute
Number of bugs found per minute of fuzzing.
Indicates efficiency of the fuzzing process in finding bugs over time.
Derived metric computed as `Total_Bugs / (Total_Time_s / 60)`.

---

## 14. Runs_per_Minute
Number of fuzzing runs executed per minute.
Derived metric computed as `Total_Runs / (Total_Time_s / 60)`.
Measures execution throughput (performance) of the fuzzing process.

---

## 15. Replays_Written
Number of replay traces generated for debugging.
Extracted from `statsFuzzing_*.csv`, summing all columns starting with `NoReplayWritten`.
Indicates how many bug-inducing traces were successfully recorded.

---

## 16. Replays_Successful
Number of replay traces that were successfully reproduced.
Confirms how many of the written replays actually reproduced the bug, validating bug accuracy.
Extracted from `statsFuzzing_*.csv`, summing all columns starting with `NoReplaySuccessful`.

---

## Metric Categories
- **Precision Metrics**: `Unique_Bugs`, `Bug_Types`, `Total_Bugs`, `Panics`, `Leaks`, `Confirmed_Replays`, `Replays_Written`, `Replays_Successful`.
- **Performance Metrics**: `Total_Runs`, `Total_Time_s`, `Rec_s`, `Ana_s`, `Rep_s`, `Bugs_per_1000_Runs`, `Bugs_per_Minute`, `Runs_per_Minute`.

These metrics provide both **effectiveness (how many bugs were found)** and **efficiency (how fast and reliably bugs were discovered)** for each fuzzing mode and test case.
