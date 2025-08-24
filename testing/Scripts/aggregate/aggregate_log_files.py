#!/usr/bin/env python3
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "runners"))
from runner_core import load_config


def aggregate_logs(test_dir, modes, log_files):
    combined_dir = os.path.join(test_dir, "combined", "logs")
    os.makedirs(combined_dir, exist_ok=True)

    for log_name in log_files:
        combined_path = os.path.join(combined_dir, log_name)
        with open(combined_path, "w", encoding="utf-8") as out_file:
            for mode in modes:
                mode_dir = os.path.join(test_dir, mode)
                if not os.path.isdir(mode_dir):
                    continue
                subdirs = [d for d in os.listdir(mode_dir) if os.path.isdir(os.path.join(mode_dir, d))]
                for sub in subdirs:
                    file_path = os.path.join(mode_dir, sub, log_name)
                    if os.path.exists(file_path):
                        out_file.write(f"\n\n### Mode: {mode} | Subdir: {sub} ###\n\n")
                        with open(file_path, "r", encoding="utf-8") as f:
                            out_file.write(f.read())
    print(f"Aggregated log files written to: {combined_dir}")


def main():
    if len(sys.argv) != 2:
        print("Usage: python3 aggregate_log_files.py <path/to/TestX_folder>")
        sys.exit(1)

    test_dir = sys.argv[1]
    if not os.path.isdir(test_dir):
        print(f"Invalid folder: {test_dir}")
        sys.exit(1)

    cfg = load_config()
    aggregate_logs(test_dir, cfg.get("modes", []), cfg.get("log_files", []))


if __name__ == "__main__":
    main()
