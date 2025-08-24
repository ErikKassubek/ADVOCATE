#!/bin/bash

echo "Verifying Go environment for Advocate..."

# Go binary in shell
echo "- Go Binary (system shell): $(which go)"
echo "- Go Version: $(go version)"
echo "- GOROOT: $GOROOT"
echo "- GOPATH: $GOPATH"


# Check system Go 
if ! go version | grep -q "go1.24.1"; then
  echo "[WARNING]: Your shell is not using Go 1.24.1. This is OK if Advocate is using its own runtime."
else
  echo "[INFO] Shell Go version is 1.24.1"
fi


# Check if patched Go binary exists
PATCHED_GO="$HOME/Projects/Go_Fuzzing_Concurrency/ADVOCATE/go-patch/bin/go"
if [ -x "$PATCHED_GO" ]; then
  echo "[INFO] Found patched Advocate Go runtime at: $PATCHED_GO"
else
  echo "[WARNING] Patched Go runtime not found at expected path"
  exit 1
fi


# Version of patched Go
PATCHED_VERSION=$($PATCHED_GO version)
echo "[INFO] Patched Go Version: $PATCHED_VERSION"

echo "... Advocate setup appears correct."
