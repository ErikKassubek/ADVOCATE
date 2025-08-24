#!/usr/bin/env python3
import os
import re
import sys
import yaml
import shutil
import subprocess
from typing import List, Dict


# ---------- Path & Config ----------
def get_project_root() -> str:
    runners_dir = os.path.dirname(os.path.abspath(__file__))      # .../testing/Scripts/runners
    testing_dir = os.path.dirname(os.path.dirname(runners_dir))   # .../testing
    return os.path.dirname(testing_dir)                           # .../


def load_config() -> Dict:
    project_root = get_project_root()
    config_path = os.path.join(project_root, "testing", "config.yaml")
    if not os.path.exists(config_path):
        raise FileNotFoundError(f"Config file not found: {config_path}")
    with open(config_path, "r") as f:
        return yaml.safe_load(f) or {}


def resolve_advocate_bin(cfg: Dict) -> str:
    project_root = get_project_root()
    rel_bin = cfg.get("advocate_bin", "./advocate/advocate")
    bin_path = os.path.normpath(os.path.join(project_root, rel_bin))
    if not os.path.isfile(bin_path) or not os.access(bin_path, os.X_OK):
        print(f"Warning: advocate binary not found or not executable at: {bin_path}")
    return bin_path


# ---------- Test Discovery ----------
_TEST_FUNC_RE = re.compile(r'^\s*func\s+(Test\w+)\s*\(', re.UNICODE)

def list_tests(test_dir: str) -> List[str]:
    """Parse all *_test.go files in test_dir and return sorted Go test function names."""
    tests: List[str] = []
    for fname in os.listdir(test_dir):
        if fname.endswith("_test.go"):
            fpath = os.path.join(test_dir, fname)
            try:
                with open(fpath, "r", encoding="utf-8", errors="ignore") as f:
                    for line in f:
                        m = _TEST_FUNC_RE.match(line)
                        if m:
                            tests.append(m.group(1))
            except Exception as e:
                print(f"Warning: failed to read {fpath}: {e}")
    return sorted(set(tests))


# ---------- Results Handling ----------
# Check results folder
def ensure_results_dir(test_dir_abs: str, test_name: str, mode: str) -> str:
    rdir = os.path.join(test_dir_abs, "results", test_name, mode)
    os.makedirs(rdir, exist_ok=True)
    return rdir


# Move results 
def move_advocate_results(test_dir_abs: str, dest_results_dir: str):
    src = os.path.join(test_dir_abs, "advocateResult")
    if not os.path.isdir(src):
        return

    for name in os.listdir(src):
        src_path = os.path.join(src, name)
        dest_path = os.path.join(dest_results_dir, name)
        if os.path.exists(dest_path):
            if os.path.isdir(dest_path):
                shutil.rmtree(dest_path)
            else:
                os.remove(dest_path)
        shutil.move(src_path, dest_path)

    try:
        os.rmdir(src)
    except OSError:
        pass

    print(f"Results moved to: {dest_results_dir}")


# Aggregation
def aggregate_logs_and_stats(test_dir_abs: str, test_name: str):
    results_test_dir = os.path.join(test_dir_abs, "results", test_name)
    combined_dir = os.path.join(results_test_dir, "combined")
    os.makedirs(combined_dir, exist_ok=True)

    root = get_project_root()
    aggregate_dir = os.path.join(root, "testing", "Scripts", "aggregate") 

    print("Aggregating log and stat files...")
    subprocess.run(["python3", os.path.join(aggregate_dir, "aggregate_log_files.py"), results_test_dir], check=False)
    subprocess.run(["python3", os.path.join(aggregate_dir, "aggregate_stat_files.py"), results_test_dir], check=False)
    print(f"Aggregation complete. Files stored in: {combined_dir}")


# Comparison
def generate_comparison(test_dir_abs: str, test_name: str):
    results_test_dir = os.path.join(test_dir_abs, "results", test_name)
    root = get_project_root()
    compare_dir = os.path.join(root, "testing", "Scripts", "compare")   

    print("Generating comparison.csv...")
    subprocess.run(["python3", os.path.join(compare_dir, "compare_one_by_all.py"), results_test_dir], check=False)
    print(f"Comparison CSV created in: {results_test_dir}/comparison.csv")


# ---------- Advocate Call ----------

# Build command
def build_advocate_cmd(advocate_bin: str, mode: str, test_dir_abs: str, test_name: str, cfg: Dict) -> List[str]:
    args = [
        advocate_bin, "fuzzing",
        "-path", test_dir_abs,
        "-exec", test_name,
        "-fuzzingMode", mode,
        "-prog", test_name,
        "-maxFuzzingRun", str(cfg.get("max_runs", 100)),
        "-timeoutFuz", str(cfg.get("timeout", 60)),
        "-timeoutRec", str(cfg.get("timeout", 60)),
        "-timeoutRep", str(cfg.get("timeout", 60)),
    ]
    if cfg.get("record_time", True):
        args.append("-time")
    if cfg.get("record_stats", True):
        args.append("-stats")
    return args


# Run Fuzzing
def run_advocate(cfg: Dict, mode: str, test_path: str, test_name: str):
    advocate_bin = resolve_advocate_bin(cfg)
    test_dir_abs = os.path.dirname(os.path.abspath(test_path))
    dest_results_dir = ensure_results_dir(test_dir_abs, test_name, mode)

    cmd = build_advocate_cmd(advocate_bin, mode, test_dir_abs, test_name, cfg)
    print(f"Running: {' '.join(cmd)}")
    subprocess.run(cmd, check=False)

    move_advocate_results(test_dir_abs, dest_results_dir)


# ---------- CLI Entrypoint ----------
def main():
    """
    Usage:
      runner_core.py all_on_all <test_dir>
      runner_core.py all_on_one <test_dir> <test_function>
      runner_core.py one_on_all <test_dir> <mode>
      runner_core.py one_on_one <test_dir> <mode> <test_function>
    """
    if len(sys.argv) < 3:
        print("Usage: runner_core.py <runner_type> <test_dir> [mode] [test_function]")
        sys.exit(1)

    cfg = load_config()
    runner_type = sys.argv[1]
    test_dir = sys.argv[2]
    if not os.path.isdir(test_dir):
        print(f"Error: test_dir not found: {test_dir}")
        sys.exit(1)

    modes = cfg.get("modes", [])
    tests = list_tests(test_dir)

    if runner_type == "all_on_all":
        for tfunc in tests:
            for mode in modes:
                run_advocate(cfg, mode, os.path.join(test_dir, tfunc), tfunc)
            aggregate_logs_and_stats(test_dir, tfunc)
            generate_comparison(test_dir, tfunc)

    elif runner_type == "all_on_one":
        if len(sys.argv) < 4:
            print("Missing test_function for all_on_one")
            sys.exit(1)
        tfunc = sys.argv[3]
        for mode in modes:
            run_advocate(cfg, mode, os.path.join(test_dir, tfunc), tfunc)
        aggregate_logs_and_stats(test_dir, tfunc)
        generate_comparison(test_dir, tfunc)

    elif runner_type == "one_on_all":
        if len(sys.argv) < 4:
            print("Missing mode for one_on_all")
            sys.exit(1)
        mode = sys.argv[3]
        for tfunc in tests:
            run_advocate(cfg, mode, os.path.join(test_dir, tfunc), tfunc)

    elif runner_type == "one_on_one":
        if len(sys.argv) < 5:
            print("Missing args: mode and test_function for one_on_one")
            sys.exit(1)
        mode = sys.argv[3]
        tfunc = sys.argv[4]
        run_advocate(cfg, mode, os.path.join(test_dir, tfunc), tfunc)

    else:
        print("Invalid runner type")
        sys.exit(1)


if __name__ == "__main__":
    main()
