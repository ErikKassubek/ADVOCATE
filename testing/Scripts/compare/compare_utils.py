#!/usr/bin/env python3
import os
import csv
import glob

def find_file(pattern: str, directory: str) -> str:
    matches = glob.glob(os.path.join(directory, pattern))
    return matches[0] if matches else None


def read_csv(path: str) -> list:
    with open(path, newline='', encoding="utf-8") as f:
        return list(csv.reader(f))


def parse_metrics(mode_dir: str) -> dict:
    metrics = {}

    # Find relevant files
    stats_all = find_file("statsAll_*.csv", mode_dir)
    stats_analysis = find_file("statsAnalysis_*.csv", mode_dir)
    stats_fuzzing = find_file("statsFuzzing_*.csv", mode_dir)
    times_total = find_file("times_total_*.csv", mode_dir)
    times_detail = find_file("times_detail_*.csv", mode_dir)

    # statsAll -> unique bugs, types
    if stats_all:
        rows = read_csv(stats_all)
        if len(rows) > 1:
            header, values = rows[0], rows[1]
            uniq_ids = [h.replace("NoUniqueDetected", "") for h, v in zip(header, values)
                        if h.startswith("NoUniqueDetected") and int(v) > 0]
            metrics["Unique_Bugs"] = len(uniq_ids)
            metrics["Bug_Types"] = ";".join(uniq_ids)

    # statsAnalysis -> total bugs, leaks, panics, confirmed replays
    if stats_analysis:
        rows = read_csv(stats_analysis)
        if len(rows) > 1:
            values = rows[1]
            metrics.update({
                "Total_Bugs": int(values[1]),
                "Leaks": int(values[2]),
                "Panics": int(values[5]),
                "Confirmed_Replays": int(values[6])
            })

    # statsFuzzing -> total runs, replays
    if stats_fuzzing:
        rows = read_csv(stats_fuzzing)
        if len(rows) > 1:
            header, values = rows[0], rows[1]
            metrics["Total_Runs"] = int(values[1])
            metrics["Replays_Written"] = sum(int(v) for h, v in zip(header, values) if h.startswith("NoReplayWritten"))
            metrics["Replays_Successful"] = sum(int(v) for h, v in zip(header, values) if h.startswith("NoReplaySuccessful"))

    # times_total -> total time
    if times_total:
        rows = read_csv(times_total)
        if len(rows) > 1 and rows[1][1]:
            metrics["Total_Time_s"] = float(rows[1][1])

    # times_detail -> breakdown
    if times_detail:
        rows = read_csv(times_detail)
        if len(rows) > 1:
            header, values = rows[0], rows[1]
            for h, v in zip(header, values):
                if h in ("Recording", "Analysis", "Replay"):
                    metrics[h[:3] + "_s"] = v

    # Derived metrics
    runs = metrics.get("Total_Runs", 0)
    bugs = metrics.get("Total_Bugs", 0)
    time_s = metrics.get("Total_Time_s", 0)
    if runs and bugs:
        metrics["Bugs_per_1000_Runs"] = round((bugs / runs) * 1000, 2)
    if time_s > 0:
        metrics["Bugs_per_Minute"] = round(bugs / (time_s / 60), 2) if bugs else 0
        metrics["Runs_per_Minute"] = round(runs / (time_s / 60), 2) if runs else 0

    return metrics


def gather_modes(results_dir: str) -> list:
    return sorted([
        d for d in os.listdir(results_dir)
        if os.path.isdir(os.path.join(results_dir, d)) and d != "combined"
    ])
