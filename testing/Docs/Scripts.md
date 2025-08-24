# Scripts Overview

The [`Scripts/`](../Scripts/) folder contains files used for running the fuzzing process, handling configurations, aggregating and comparing the results.

```
├── run.py
├── config.yaml  
├── Scripts/
  ├── runners/
  │ ├── runner_core.py
  │ ├── run_all_on_all.sh 
  │ ├── run_all_on_one.sh 
  │ ├── run_one_on_all.sh 
  │ ├── run_one_on_one.sh 
  │ 
  ├── setup/ 
  │ ├── check_setup.sh 
  │ ├── clean_results.sh 
  │ 
  ├── aggregate/
  │ ├── aggregate_log_files.py
  │ ├── aggregate_stat_files.py 
  │ 
  ├── compare/
  │ ├── compare_all_by_one.py
  │ ├── compare_one_by_all.py
  │ ├── compare_utils.py
```

## Descriptions

### [`config.yaml`](../config.yaml)
The central configuration defines the Advocate binary path (advocate_bin), the available and to be tested fuzzing modes, execution limits like max_runs and timeout, as well as additional settings like log_files and stat_prefixes to be considered for aggregation and comparison.


### [`run.py`](../run.py)
Provides an interactive interface for selecting the run type (all modes/all tests, all modes/one test, etc.) and the test case and/or mode when required.

It delegates the actual execution to the appropriate runner script in for a clean separation between user interaction, execution logic, and post-processing.

### [`setup/`](../Scripts/setup/)
- [`check_setup.sh`](../Scripts/setup/check_setup.sh): Verifies that system and Advocate environment are correctly set up, by checking the system Go binary and version, presence of the patched Advocate Go runtime (`go-patch`) and that it is executable, version of the patched Go runtime

- [`clean_results.sh`](../Scripts/setup/clean_results.sh): Removes all result artifacts created by fuzzing runs. You can customize it by modifying the target folder path, which is recursively searched for artifacts to delete.


### [`runners/`](../Scripts/runners/)
Manages running Advocate in different modes/tests combinations.
[`runner_core.py`](../Scripts/runners/runner_core.py) is the core module that handles configuration loading (config.yaml), test case listing, and building the command to execute the Advocate binary. 

There are four thin wrappers that call runner_core.py with the correct parameters representing the available modes/ testing combinations.

1. [`run_all_on_all.sh`](../Scripts/runners/run_all_on_all.sh): All modes on all test cases.

2. [`run_all_on_one.sh`](../Scripts/runners/run_all_on_one.sh): All modes on a single test case.

3. [`run_one_on_all.sh`](../Scripts/runners/run_one_on_all.sh): One selected mode across all tests.

4. [`run_one_on_one.sh`](../Scripts/runners/run_one_on_one.sh): One selected mode on one selected test.

These scripts are triggered by [`run.py`](../run.py), which presents an interactive selection menu.

As **`all_on_all`** and **`all_on_one`** are using all modes defined in the `config` for fuzzing tests,  result aggregation and comparison generation is triggered by executing the script [`compare_one_by_all.py`](../Scripts/aggregate/aggregate_log_files.py) 



### [`aggregate/`](../Scripts/aggregate/)
- [`aggregate_log_files.py`](../Scripts/aggregate/aggregate_log_files.py): Aggregates log files into a single combined file for further analysis.

- [`aggregate_stat_files.py`](../Scripts/aggregate/aggregate_stat_files.py): Aggregates statistical data into a combined file to simplify comparison between runs.

These scripts reuse load_config() from runner_core.py to stay in sync with the central configuration.



### [`compare/`](../Scripts/compare/)
- [`compare_utils.py`](../Scripts/compare/compare_utils.py): Provides reusable parsing utilities to extract metrics (bugs, runtime, performance indicators) from Advocate results

- [`compare_one_by_all.py`](../Scripts/compare/compare_one_by_all.py): Uses these utilities to generate a **comparison.csv** file summarizing all metrics for each mode in a single test run.

- [`compare_all_by_one.py`](../Scripts/compare/compare_all_by_one.py): Enables cross-test analysis, allowing users to interactively choose a metric and see a pivoted comparison of that metric across all tests and modes. (Requires existing **comparison.csv**)
