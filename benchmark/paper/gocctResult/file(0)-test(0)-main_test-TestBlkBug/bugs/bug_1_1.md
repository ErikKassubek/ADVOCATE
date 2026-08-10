# Bug: A07 - Actual Non-Cyclic Blocking Bug

During the execution, a blocking bug was detected.
This means, there is a routine that is blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestBlkBug
- File: /home/erik/Arbeit/GoCCT/benchmark/paper/main_test.go
- Trace: gocctTrace_2

## Bug Elements

The elements involved in the found bug are located at the following positions:

###  Channel: Send

-> /home/erik/Arbeit/GoCCT/benchmark/paper/main_test.go#12

```go 2 
 3 import (
 4 	"fmt"
 5 	"testing"
 6 	"time"
 7 )
 8 
 9 func TestBlkBug(_ *testing.T) {
10 	ch := make(chan int)
11 	go func() {
12 		ch <- calc()                    // <================= 
13 	}()
14 
15 	select {
16 	case x := <-ch:
17 		fmt.Println(x)
18 	case <-time.After(1 * time.Second):
19 		fmt.Println("Timeout")
20 	}
21 }
22 
...
```

