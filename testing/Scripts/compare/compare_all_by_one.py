#!/usr/bin/env python3
import os
import sys
import pandas as pd


def compare_all(root_dir: str):
    rows = []
    for test in os.listdir(root_dir):
        comp_path = os.path.join(root_dir, test, "comparison.csv")
        if os.path.isfile(comp_path):
            df = pd.read_csv(comp_path)
            for _, row in df.iterrows():
                rows.append({"Test": test, **row.to_dict()})

    if not rows:
        print("No comparison.csv found in subfolders.")
        return

    df_all = pd.DataFrame(rows)
    metrics = [c for c in df_all.columns if c not in ("Test", "Mode")]

    print("\nAvailable metrics:")
    for i, m in enumerate(metrics, 1):
        print(f"{i}) {m}")
    choice = input("Select metric by number: ").strip()
    if not choice.isdigit() or not (1 <= int(choice) <= len(metrics)):
        print("Invalid choice.")
        return

    metric = metrics[int(choice) - 1]
    pivot = df_all.pivot(index="Test", columns="Mode", values=metric)
    print(f"\n=== {metric} COMPARISON ===")
    print(pivot.fillna("-").to_string())

    out_path = os.path.join(root_dir, f"comparison_pivot_{metric}.csv")
    pivot.to_csv(out_path)
    print(f"\nSaved: {out_path}")


def main():
    if len(sys.argv) != 2:
        print("Usage: python3 compare_all.py <root_results_dir>")
        sys.exit(1)
    root_dir = sys.argv[1]
    if not os.path.isdir(root_dir):
        print(f"Invalid folder: {root_dir}")
        sys.exit(1)
    compare_all(root_dir)


if __name__ == "__main__":
    main()
