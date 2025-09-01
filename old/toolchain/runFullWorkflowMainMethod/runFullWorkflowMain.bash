while [[ $# -gt 0 ]]; do
	key="$1"
	case $key in
	-a | --advocate)
		pathToAdvocate="$2"
		shift
		shift
		;;
	-f | --main-file)
		file="$2"
		shift
		shift
		;;
	*)
		shift
		;;
	esac
done

pathToPatchedGoRuntime="$pathToAdvocate/goPatch/bin/go"
pathToGoRoot="$pathToAdvocate/goPatch"
pathToOverheadInserter="$pathToAdvocate/toolchain/overHeadInserter/inserter"
pathToOverheadRemover="$pathToAdvocate/toolchain/overHeadRemover/remover"
pathToAnalyzer="$pathToAdvocate/analyzer/analyzer"

if [ -z "$pathToAdvocate" ]; then
	echo "Path to advocate is empty"
	exit 1
fi
if [ -z "$file" ]; then
    echo "Main file is empty"
    exit 1
fi

dir=$(dirname "$file")
cd $dir
echo "In directory: $dir"
echo "export GOROOT=$pathToGoRoot"
export GOROOT=$pathToGoRoot
echo "Goroot exported"
echo "Remove Overhead just in case"
$pathToOverheadRemover -f $file
echo "Add Overhead"
echo "$pathToOverheadInserter -f $file"
$pathToOverheadInserter -f $file
if [ $? -ne 0 ]; then
	echo "Error in adding overhead"
	exit 1
fi
echo "Run test"
echo "$pathToPatchedGoRuntime run $file"
$pathToPatchedGoRuntime run $file
if [ $? -ne 0 ]; then
	echo "Remove Overhead"
	$pathToOverheadRemover -f $file
	echo "Error in running test, therefor overhead removed and full workflow stopped."
	exit 1
fi
echo "Remove Overhead"
echo "$pathToOverheadRemover -f $file"
$pathToOverheadRemover -f $file
echo "Apply analyzer"
echo "$pathToAnalyzer -t $dir/advocateTrace"
$pathToAnalyzer -t "$dir/advocateTrace"
rewritten_traces=$(find "$dir" -type d -name "rewritten_trace*")
for trace in $rewritten_traces; do
	rtracenum=$(echo $trace | grep -o '[0-9]*$')
	echo "Apply reorder overhead"
	echo $pathToOverheadInserter -f $file -r true -n "$rtracenum"
	$pathToOverheadInserter -f $file -r true -n "$rtracenum"
	echo "$pathToPatchedGoRuntime run $file"
	$pathToPatchedGoRuntime run $file 2>&1 | tee -a "$trace/reorder_output.txt"
	echo "Remove reorder overhead"
	echo "$pathToOverheadRemover -f $file"
	$pathToOverheadRemover -f $file
done
unset GOROOT
