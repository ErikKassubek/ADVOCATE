# Setup Advocate & Run Samples

This guide walks you through installing, building, and running a basic test with Advocate.  
It ensures your setup is working correctly before executing automated fuzzing workflows.

For further information see: [How To Use Advocate](https://github.com/ErikKassubek/ADVOCATE/blob/main/doc/usage.md)


## Install Advocate

### Prerequisites
- Linux OS or WSL
- Go Version 1.24.1

--- 

### Step 1: Update System

```bash
sudo apt update
sudo apt upgrade
```

---

### Step 2: Install Go 1.24.1
```bash
cd ~/Downloads 
curl -LO https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
tar -xzf go1.24.1.linux-amd64.tar.gz
mkdir -p ~/Tools
mv go ~/Tools/go-runtime
```

---

### Step 3: Set Go Environment
```bash
export GOROOT=$HOME/Tools/go-runtime
export PATH=$GOROOT/bin:$PATH
source ~/.bashrc 
```

---

### 4. Verify Version
```bash
go version  # go version go1.24.2 linux/amd64
which go    # /home/USER/Tools/go-runtime/bin/go
```

---

### 5. Clone & Build Advocate
```bash
# Clone git
git clone https://github.com/ErikKassubek/ADVOCATE.git
cd ADVOCATE

# Build patched go runtime (using ~/Tools/go-runtime Go 1.24.1)
# Patch version lives in /home/USER/Projects/Go_Fuzzing_Concurrency/ADVOCATE/go-patch/bin/go
cd go-patch/src
./make.bash

# Add to path
export PATH=$HOME/Projects/Go_Fuzzing_Concurrency/ADVOCATE/go-patch/bin:$PATH

# (To Remove)
# export PATH=$(echo $PATH | sed -e "s|$HOME/Projects/Go_Fuzzing_Concurrency/ADVOCATE/go-patch/bin:||") 

# Build CLI
cd ../../advocate
go build
```

---


### 6. Test Advocate Sample
Run in /ADVOCATE/advocate:
```bash
./advocate fuzzing \
-path ~/Projects/Go_Fuzzing_Concurrency/ADVOCATE/examples/deadlocks/ \
-exec TestBasicLockdep \
-fuzzingMode GoPie \
-prog deadlock_test \
```