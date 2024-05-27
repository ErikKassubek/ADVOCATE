

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
    -t|--test-name)
      testName="$2"
      shift
      shift
      ;;
    -p |--package)
      package="$2"
      shift
      shift
      ;;
    -tf|--test-file)
      file="$2"
      shift
      shift
      ;;
    *)
      shift
      ;;
  esac
done

# previous command
#./unitTestFullWorkflow.bash -p /home/mario/Desktop/thesis/ADVOCATE/go-patch/bin/go -g /home/mario/Desktop/thesis/ADVOCATE/go-patch -i /home/mario/Desktop/thesis/ADVOCATE/toolchain/unitTestOverheadInserter/unitTestOverheadInserter -r /home/mario/Desktop/thesis/ADVOCATE/toolchain/unitTestOverheadRemover/unitTestOverheadRemover -a /home/mario/Desktop/thesis/ADVOCATE/analyzer/analyzer -f ~/Desktop/fullMod -t TestSomething -package module/path -tf /home/mario/Desktop/fullMod/module/path/some_test.go 

#Initialize Variables
pathToPatchedGoRuntime="$pathToAdvocate/go-patch/bin/go"
pathToGoRoot="$pathToAdvocate/go-patch"
pathToOverheadInserter="$pathToAdvocate/toolchain/unitTestOverheadInserter/unitTestOverheadInserter"
pathToOverheadRemover="$pathToAdvocate/toolchain/unitTestOverheadRemover/unitTestOverheadRemover"
pathToAnalyzer="$pathToAdvocate/analyzer/analyzer"



if [ -z "$pathToAdvocate" ]; then
  echo "Path to advocate is empty"
fi
if [ -z "$dir" ]; then
  echo "Directory is empty"
fi
if [ -z "$testName" ]; then
  echo "Test name is empty"
fi
if [ -z "$package" ]; then
  echo "Package is empty"
fi
if [ -z "$file" ]; then
  echo "Test file is empty"
fi



cd "$dir"
echo  "In directory: $dir"
export GOROOT=$pathToGoRoot
echo "Goroot exported"
#Remove Overhead just in case
echo "Remove Overhead just in case"
#echo "$pathToOverheadRemover -f $file -t $testName"
$pathToOverheadRemover -f $file -t $testName
#Add Overhead
echo "Add Overhead"
$pathToOverheadInserter -f $file -t $testName
##Run test
echo "Run test"
echo "$pathToPatchedGoRuntime test -count=1 -run=$testName ./$package"
$pathToPatchedGoRuntime test -count=1 -run=$testName "./$package"
##Remove Overhead
echo "Remove Overhead"
$pathToOverheadRemover -f $file -t $testName
#Run Analyzer
$pathToAnalyzer -t "$dir/$package/advocateTrace"
#Loop through every rewritten traces
rewritten_traces=$(find "./$package" -type d -name "rewritten_trace*")
rtracenum=1
for trace in $rewritten_traces; do
  ## Apply reorder overhead
  $pathToOverheadInserter -f $file -t $testName -r true -n $rtracenum
  ## Run test
  echo "Run reordered test"
  $pathToPatchedGoRuntime test -count=1 -run=$testName "./$package"
  ## Remove reorder overhead
  $pathToOverheadRemover -f $file -t $testName
  rtracenum=$((rtracenum+1))
done
unset GOROOT