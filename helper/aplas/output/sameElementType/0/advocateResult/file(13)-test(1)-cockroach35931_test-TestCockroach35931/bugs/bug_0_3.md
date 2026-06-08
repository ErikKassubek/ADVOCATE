# Bug: A08 - Actual Cyclic Deadlock

During the execution, a deadlock was detected.
This means, there is are routine that are cyclicaly blocked, and there is not possibility of it being unblocked in the future

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCockroach35931
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/0/blocking/cockroach/35931/cockroach35931_test.go
- Trace: advocateTrace_14

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Select:
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/0/blocking/cockroach/2448/cockroach2448_test.go#59
```go
48 ...
49 
50 }
51 func (s *state) handleWriteResponse() {
52 	s.processCommittedEntry()
53 }
54 
55 func (s *state) processCommittedEntry() {
56 	s.sendEvent(&EventMembershipChangeCommitted{
57 		Callback: func() {
58 			select {
59 			case s.callbackChan <- func() { // Waiting for callbackChan consumption           // <-------
60 				time.Sleep(time.Nanosecond)
61 			}:
62 			case <-s.stopper.ShouldStop():
63 			}
64 		},
65 	})
66 }
67 
68 type Store struct {
69 	multiraft *MultiRaft
70 
71 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/0/blocking/cockroach/2448/cockroach2448_test.go#42
```go
31 ...
32 
33 }
34 
35 type state struct {
36 	*MultiRaft
37 }
38 
39 func (s *state) start() {
40 	for {
41 		select {
42 		case <-s.stopper.ShouldStop():           // <-------
43 			return
44 		case cb := <-s.callbackChan:
45 			cb()
46 		default:
47 			s.handleWriteResponse()
48 		}
49 	}
50 }
51 func (s *state) handleWriteResponse() {
52 	s.processCommittedEntry()
53 
54 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/sameElementType/0/blocking/cockroach/35073/cockroach35073_test.go#49
```go
38 ...
39 
40 	}
41 }
42 
43 func (rc *RowChannel) Push() ConsumerStatus {
44 	consumerStatus := ConsumerStatus(
45 		atomic.LoadUint32((*uint32)(&rc.consumerStatus)))
46 	switch consumerStatus {
47 	case NeedMoreRows:
48 		rc.dataChan <- RowChannelMsg(0)
49 	case DrainRequested:           // <-------
50 	case ConsumerClosed:
51 	}
52 	return consumerStatus
53 }
54 
55 func (rc *RowChannel) InitWithNumSenders() {
56 	rc.initWithBufSizeAndNumSenders(rowChannelBufSize)
57 }
58 
59 func (rc *RowChannel) initWithBufSizeAndNumSenders(chanBufSize int) {
60 
61 ...
```


## Replay
**Replaying was not run**.

