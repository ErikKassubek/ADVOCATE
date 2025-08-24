#!/usr/bin/env bash
set -euo pipefail

echo "Cleaning all advocateResult, fuzzingTraces, results folders and output.log files..."

# Delete all created dirs / files in ./Examples
find ./Examples -type d -name "advocateResult" -exec rm -rf {} +
find ./Examples -type d -name "advocateTrace" -exec rm -rf {} +
find ./Examples -type d -name "fuzzingTraces" -exec rm -rf {} +
find ./Examples -type d -name "results" -exec rm -rf {} +
find ./Examples -type f -name "output.log" -exec rm -f {} +

echo "Cleanup complete."
