#!/usr/bin/env python3
import os
import sys
import subprocess

# Helper functions
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "Scripts/runners"))
from Scripts.runners.runner_core import load_config, list_tests


# Show options for selection
def select_from_list(title, options):
    print(title)
    for idx, opt in enumerate(options, start=1):
        print(f"{idx}) {opt}")
    while True:
        choice = input("#? ").strip()
        if choice.isdigit() and 1 <= int(choice) <= len(options):
            return options[int(choice) - 1]
        print("Invalid choice, try again.")


# User selection
def main():
    if len(sys.argv) < 2:
        print("Usage: run.py <path_to_test_dir>")
        sys.exit(1)

    # Get test dir & load test cases
    test_dir = sys.argv[1]
    if not os.path.isdir(test_dir):
        print(f"Error: {test_dir} is not a directory.")
        sys.exit(1)

    tests = list_tests(test_dir)
    if not tests:
        print(f"No test cases found in {test_dir}")
        sys.exit(1)

    # Load config with available modes
    cfg = load_config()
    modes = cfg.get("modes", [])

    # Choose Runner
    print("========== Advocate Fuzzing Runner ==========")
    print("Select option:")
    print("  1) Run ALL modes on ALL test cases")
    print("  2) Run ALL modes on ONE selected test case")
    print("  3) Run ONE selected mode on ALL test cases")
    print("  4) Run ONE selected mode on ONE selected test case")
    print("  5) Exit")
    choice = input("Choice [1-5]: ").strip()

    if choice == "5":
        print("Exiting...")
        sys.exit(0)

    selected_test = None
    selected_mode = None

    # Choose Fuzzing Mode
    if choice in ("3", "4"):
        selected_mode = select_from_list("Select fuzz mode:", modes)

    # Choose Test Case
    if choice in ("2", "4"):
        selected_test = select_from_list("Available Tests:", tests)

    # Call selected runner script
    base_dir = os.path.dirname(__file__)
    runners_dir = os.path.join(base_dir, "Scripts/runners")

    if choice == "1":
        runner = os.path.join(runners_dir, "run_all_on_all.sh")
        args = [runner, test_dir]
    elif choice == "2":
        runner = os.path.join(runners_dir, "run_all_on_one.sh")
        args = [runner, test_dir, selected_test]
    elif choice == "3":
        runner = os.path.join(runners_dir, "run_one_on_all.sh")
        args = [runner, test_dir, selected_mode]
    elif choice == "4":
        runner = os.path.join(runners_dir, "run_one_on_one.sh")
        args = [runner, test_dir, selected_mode, selected_test]
    else:
        print("Invalid choice")
        sys.exit(1)

    print(f"Running: {' '.join(args)}")
    subprocess.run(args)


if __name__ == "__main__":
    main()
