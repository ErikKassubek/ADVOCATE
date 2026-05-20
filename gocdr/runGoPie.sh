#!/bin/bash

# Define your list of values
list=(2 5 10 25 50 75 100)

# Loop through the list and run ./gocdr with each item
for i in "${list[@]}"; do
    echo "Running $i"
    ./gocdr fuzzing -path ~/Uni/gocdr/examples/goker/gobench -mode GoPie -timeoutRec 20 -maxFuzzingRuns $i -noInfo -noProgress

done
