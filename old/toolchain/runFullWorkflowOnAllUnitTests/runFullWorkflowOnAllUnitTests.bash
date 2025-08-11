#!/bin/bash
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    -a|--advocate)
      pathToAdvocate="$2"
      shift
      shift
      ;;
    -f|--folder)
      dir="$2"
      shift
      shift
      ;;
    -m|--modulemode)
      modulemode="$2"
      shift
      shift
      ;;
    *)
      shift
      ;;
  esac
done

if [ -z "$pathToAdvocate" ]; then
  echo "Path to advocate is empty"
  exit 1
fi
if [ -z "$dir" ]; then
  echo "Directory is empty"
  exit 1
fi

pathToPatchedGoRuntime="$pathToAdvocate/goPatch/bin/go"
pathToGoRoot="$pathToAdvocate/goPatch"
pathToOverheadInserter="$pathToAdvocate/toolchain/unitTestOverheadInserter/unitTestOverheadInserter"
pathToOverheadRemover="$pathToAdvocate/toolchain/unitTestOverheadRemover/unitTestOverheadRemover"
pathToAnalyzer="$pathToAdvocate/analyzer/analyzer"
pathToFullWorkflowExecutor="$pathToAdvocate/toolchain/unitTestFullWorkflow/unitTestFullWorkflow.bash"

cd "$dir"
echo  "In directory: $dir"
mkdir -p advocateResult

test_files=$(find "$dir" -name "*_test.go")
total_files=$(echo "$test_files" | wc -l)
current_file=1
skipped_tests=0
attempted_tests=0
for file in $test_files; do
    echo "Progress: $current_file/$total_files"
    echo "Processing file: $file"
    package_path=$(dirname "$file")
    test_functions=$(grep -oE ".*Test.*\(.*testing\.T\)" $file | sed 's/(.*\*testing\.T.*)//' | sed 's/func //')
    for test_func in $test_functions; do
        attempted_tests=$((attempted_tests+1))
        packageName=$(basename "$package_path")
        fileName=$(basename "$file")
        echo "Running full workflow for test: $test_func in package: $package_path in file: $file"
        adjustedPackagePath=$(echo "$package_path" | sed "s|$dir||g")
        directoryName="advocateResult/file($current_file)-test($attempted_tests)-$fileName-$test_func"
        mkdir -p $directoryName
        if [ "$modulemode" == "true" ]; then
            $pathToFullWorkflowExecutor -a $pathToAdvocate -p $adjustedPackagePath -m true -f $dir -tf $file -t $test_func &> $directoryName/output.txt
        else
            $pathToFullWorkflowExecutor -a $pathToAdvocate -p $adjustedPackagePath -f $dir -tf $file -t $test_func &> $directoryName/output.txt
        fi
        if [ $? -ne 0 ]; then
            echo "File $current_file with Test $attempted_tests failed, check output.txt for more information. Skipping..."
            skipped_tests=$((skipped_tests+1))
            continue
        fi
        mv $package_path/advocateTrace $directoryName
        mv $package_path/results_machine.log $directoryName
        mv $package_path/results_readable.log $directoryName
        mv $package_path/times.log $directoryName
        mv $package_path/rewritten_trace* $directoryName 2>/dev/null
        mv ./advocateCommand.log $directoryName
    done
    current_file=$((current_file+1))
done
echo "Generate Bug Reports"
echo "$pathToAdvocate/toolchain/generateBugReportsFromAdvocateResult/generateBugReports -a $pathToAdvocate -f $dir/advocateResult"
$pathToAdvocate/toolchain/generateBugReportsFromAdvocateResult/generateBugReports -a $pathToAdvocate -f $dir/advocateResult
echo "Check for untriggered selects"
$pathToAnalyzer -o -R $dir/advocateResult -P $dir
echo "Finished fullworkflow for all tests"
echo "Attempted tests: $attempted_tests"
echo "Skipped tests: $skipped_tests"
