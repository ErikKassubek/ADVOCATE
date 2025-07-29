#!/bin/bash

# Define your list of values
list=(2 5 10 25 50 75 100)

# Loop through the list and run ./advocate with each item
for i in "${list[@]}"; do
    echo "Running $i"
    ./advocate fuzzing -path ~/Uni/Advocate/examples/goker/gobenchplus/ -fuzzingMode GoPie+ -timeoutRec 20 -maxFuzzingRuns $i -noInfo -noProgress -panic

done
