#!/usr/bin/env python3
import os
import sys
import csv
from compare_utils import gather_modes, parse_metrics


def generate_comparison(results_dir: str):
    modes = gather_modes(results_dir)
    rows = []
    for mode in modes:
        metrics = parse_metrics(os.path.join(results_dir, mode))
        rows.append({"Mode": mode, **metrics})

    # Dynamic header
    header = sorted({k for row in rows for k in row})

    out_path = os.path.join(results_dir, "comparison.csv")
    with open(out_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=header)
        writer.writeheader()
        for row in rows:
            writer.writerow(row)
    print(f"comparison.csv written to {out_path}")


def main():
    if len(sys.argv) != 2:
        print("Usage: python3 generate_comparison.py <results_dir>")
        sys.exit(1)
    results_dir = sys.argv[1]
    if not os.path.isdir(results_dir):
        print(f"Invalid folder: {results_dir}")
        sys.exit(1)
    generate_comparison(results_dir)


if __name__ == "__main__":
    main()
