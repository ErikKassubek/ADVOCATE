#!/usr/bin/env python3
import os
import sys
import csv
import glob

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "runners"))
from runner_core import load_config


def find_stat_files(test_dir, stat_prefix):
    pattern = os.path.join(test_dir, "*", f"{stat_prefix}_*.csv")
    return sorted(glob.glob(pattern))


def extract_mode_from_path(path):
    return os.path.basename(os.path.dirname(path))


def aggregate_stats(test_dir, stat_prefixes):
    combined_dir = os.path.join(test_dir, "combined", "stats")
    os.makedirs(combined_dir, exist_ok=True)

    for prefix in stat_prefixes:
        stat_files = find_stat_files(test_dir, prefix)
        if not stat_files:
            continue

        output_file = os.path.join(combined_dir, os.path.basename(stat_files[0]))
        with open(output_file, "w", newline='') as out_csv:
            writer = None
            for f in stat_files:
                with open(f, newline='') as in_csv:
                    reader = csv.reader(in_csv)
                    rows = list(reader)
                    if not rows:
                        continue
                    mode = extract_mode_from_path(f)
                    if writer is None:
                        header = ["Mode"] + rows[0]
                        writer = csv.writer(out_csv)
                        writer.writerow(header)
                    for row in rows[1:]:
                        writer.writerow([mode] + row)
    print(f"Aggregated stat files written to: {combined_dir}")


def main():
    if len(sys.argv) != 2:
        print("Usage: python3 aggregate_stat_files.py <path/to/TestX_folder>")
        sys.exit(1)

    test_dir = sys.argv[1]
    if not os.path.isdir(test_dir):
        print(f"Invalid folder: {test_dir}")
        sys.exit(1)

    cfg = load_config()
    aggregate_stats(test_dir, cfg.get("stat_prefixes", []))


if __name__ == "__main__":
    main()
