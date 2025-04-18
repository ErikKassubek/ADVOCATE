The main fuzzingloop is implemented as follows.

The fuzzing contains a queue with all the mutations to run.
When the program or test is run for the first time, it will be run in the normal recording mode of the toolchain. Otherwise it will pop a mutation from the queue and run this mutation. This is done by storing the relevant information in a file called`fuzzingData.log` and adding a specific header to the program/test. This will run the mutation (see [Running a mutation](#running-a-mutation)) and record the trace for this mutation.

Then the analyzer will be applied to the recorded trace to find potential bugs. If replays are possible, they will be performed here as well.

Afterwards the fuzzing will parse the internal trace and calculate all values required to determine whether the run was interesting and if so, how many and which new mutations should be created.

There are two ways to create new mutations. One is an improved version of [GFuzz](#GFuzz). The [other one](#Path\ expansion) specifically reorders operations in such a way, that it may result in previously not executed program parts being executed now. It mainly looks at Once, TryLock and Channel order.

Each new mutation is either created by the improved GFuzz or by the path expansion, not by bot. These new mutations are then added to the queue to be executed later

This loop is repeated until the mutation queue is empty. Additionally a maximum number of runs or a maximum time can be set.


## GFuzz

### Determine whether the run was interesting
- A run is interesting, if one of the following conditions is met. The underlying information need to be stored in a file for the following runs.
  - The run contains a new pair of channel operations
    - All pairs of channel operations (send-recv) must be stored
  - If an operation pair's execution counter changes significantly from previous order.
    - For each operation pair determine the average number of executions per run
    - Determine a run to be interesting, if the number diverges from this average by at least 50%
    - Details changed from original GFuzz paper, because definition in GFuzz paper makes no sense.
  - If a new channel operations (creation, close, not close) is triggered for the first time
    - We must be able to identify each channel
      - Add a trace element on create -> channels can be identified by path of create and channel ID
    - We must store all channels ever created
    - We must store for all channels that have ever been created, whether they have been ever closed/not closed or both
  - If a buffered channel gets a larger maximum fullness than in all previous executions
    - For each channel we must store the maximum fullness over all runs
  - Additionally (not in the original GFuzz), we determine a run interesting if
  a select case, that has never been selected before is selected

### Determine the score
- For the base GFuzz, we need to extract the following information from the trace:
  - CountChOpPair_i: For each pair i of send/receive, how often was it executed
  - CreateCh: How many distinct channels have been created
    - Can be determined based on channel id
  - CloseCh: Number of closed channel
    - Count close operations
  - MaxChBufFull: Maximum fullness for each buffer
    - Each buffered channel info in the trace contains the current qSize. Pass all send and get the biggest
- With those values it is possible to determine the score
- Later this should be extended based on information from the happens before

The score determination is extended by information from the HB relations.
For now we extend it by increasing the the number of mutations for runs,
in which the HB analysis indicates for multiple selects, that they have
multiple possible communication partners. For this, we calculate the number
of possible communication partners for each select to get the SelPosPartner value. The calculation for the score is
then extended as
$$score = \sum\log_2 CountChOpPair + 10 \cdot CreateCh + 10 \cdot CloseCh + 10 \cdot \sum MaxChBufFull + a \cdot \sum SelPosPartner$$
where $a$ is a scaling factor that is still to be determined by experiments (currently also set to 10).


### Creating mutations
From the score of an interesting mutation we get the number of mutations to create
by calculating $$mut = \lceil 5 * score/maxScore \rceil.$$

For each of the mutations we do the following based on the recorded run.
We traverse all selects. We then decide whether the select should flip its
preferred case (see [flip probability](#flip-probability)). If the select will
flip its preferred case, it will chose one of the cases randomly. In the
original GFuzz each case seems to have the same probability of being chosen.
In our implementation we change this. The HB analysis calculated which of the
cases had a possible
communication partner in the last run. Those cases are given a higher
likelihood to be chosen (currently factor two, may be changed based on experimental
results). The chosen case in the last run will never be chosen.
The newly created mutations are then stored in a queue to be run.

### Flip Probability
The flip probability is the probability that a single select in the fuzzingData will change its preferred case compared to the previous mutation. If its set to high, the mutation mechanism basically becomes completely random. If its to small, the program will result in the same mutation being created over and over again. For now the probability is calculated as $$P = max(0.1, 1 - (1 - 0.99)^{1/numSel})$$ where $numSel$ is the number of selects in the previous mutation. This is selected in such a way, that the probability of at least one of the selects to flip its preferred case is at least $99\%$, but the probability for each individual select to get flipped is at least $10\%$. This may be changed based on experimental results.

### Running a mutation
The `fuzzingData.log` contains for each select a list the preferred cases.
The selects are identified by its position in the code base. When initializing the mutation run, the file is read in and stored in a `map[string][]int` `fuzzingSelectData` with an entry per select containing the list of preferred indexes. Additionally there is a map `fuzzingSelectDataIndex: map[string]int`, storing for each of the select the index of the next preferred case in `fuzzingSelectData`. Then a select is supposed to be executed, assuming it is not an internal select, the preferred case is retrieved via `fuzzingSelectData[selPos][fuzzingSelectDataIndex[selPos]]` followed by `fuzzingSelectDataIndex[selPos]++`.

The select is implemented in the `runtime/select.go:selctgo` function. This
function is split into two parts. On of them is the modified version and is called if the fuzzing is active.The other part is the original select and
is called if fuzzing is not active or if the fuzzing
select was terminated by a time out.

```go
func selectgo(cas0 *scase, order0 *uint16, pc0 *uintptr, nsends, nrecvs int, block bool) (int, bool) {
	fuzzingEnabled, fuzzingIndex, fuzzingTimeout := AdvocateFuzzingGetPreferredCase(2)
	if fuzzingEnabled {
		if ok, i, b := fuzzingSelect(cas0, order0, pc0, nsends, nrecvs, block, fuzzingIndex, fuzzingTimeout); ok {
			return i, b
		}
	}

	return originalSelect(cas0, order0, pc0, nsends, nrecvs, block)
}
```

The select in its original implementation contains 3 passes. In the first pass the select loops over all cases and checks whether on of them has a
possible already waiting partner. If this is the case, it is immediately executed and the select finishes. If not, and there is a default case, it is executed.\
If there is no default case and non of the cases could be executed, all case operations are enqueued into the queues of the corresponding channels.\
Then the routine is parked using `gopark` until it is woken up by another routine that wants to communicate with one of the channels in the cases.\
The communication is then executed.\
In the third pass, all case operations that have not been executed are removed from there respective queues.

The implementation for the fuzzing is changed in the following way:\
For the first pass, only the case where the internal `casi` is equal to the `fuzzingIndex` is checked for possible communication partners.
If this case can directly communicate, it is executed and the select is done.
If the default case is selected as the default cause it is directly executed.\
If neither of those happened, the preferred case is enqueued into its waiting queue. Then a modified park
function is called.

```go
func goparkWithTimeout(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceReason traceBlockReason, traceskip int, timeout int64) {
	mp := acquirem()
	gp := mp.curg

	// Setup timer if timeout is non-zero
	if timeout > 0 {
		go func() {
			sleep(float64(timeout))
			goready(gp, traceskip)
		}()
	}

	// Run the original gopark logic
  ...
```
This part method will first start a go routine with a timer. If the timer has run out, it will wake up the routine regardless of whether it has found a partner. Then the normal park function is run. If the routine was woken up because the enqueued channel found a partner, the select continues as in the unmodified version. If it was woken up by the timeout, it will remove the enqueued channel operation, do some clean up and return from the modified select returning $ok = false$. In this case the unmodified select will be run as can be seen in the `selectgo` func above.

### Improvement over original GFuzz
The improvements over GFuzz are mainly in finding interesting runs faster and
therefore needing to run fewer runs.
The main improvement is, as always, that
we do not need to run each mutation multiple times, hoping we run into a
bug. Running a mutation once (or a small number of times) is enough. This
also means, that bugs, which are theoretically possible (and can maybe even be replayed),
but which are so unlikely, that they are not detected by simply running the
program a limited amount of runs, can be detected.

GFuzz only mutates the selects. But for some reason, neither the decision
whether a run is interesting, nor the calculation of the number of mutations
depend on the selects in the recorded code. But it should be obvious, that the
number of selects, and especially
the number of select cases with possible partners influence how likely a
mutation of selects will result in new paths being explored.

In the original GFuzz, the two following examples, assuming in both the
communication on `x` is executed, result in the same
decision about whether the runs are are interesting and in the same number
of mutations (given that the rest of the program is identical)
```go
a := make(chan int)
x := make(chan int)

go func() {
  x <- 1
}

go func() {
  a <- 1
}


<-x
// do stuff
```

```go
a := make(chan int)
x := make(chan int)

go func() {
  x <- 1
}

go func() {
  a <- 1
}

select {
  case <-x:
    // do stuff
  case <-a:
    // potential bug
  ...
}
```
But it should be obvious, that a mutation of the second program
is more useful than a mutation of the first problem. By including the
number of select cases with possible partners, we are more likely to create
mutations which lead to new and interesting runs.

Additionally, the changes in the calculation of the score and the
calculation of the mutations may result in interesting runs been found faster
and preventing to many mutations that do not change the actual mutation from
being created.
```go
go func() {
  a <- 1
}()

go func() {
  z <- 1
}()

select {
  case <-a:
    //does not contain bug
  case <-b:
    //does not contain bug
  ...
  case <- z:
    // does contain bug
  default:
    // does not contain the bug
}
```
This run contains a select with many different cases. Lets assume, that based
on the HB analysis only the cases with channel `a` and channel `z` actually
have a partner and that the case with channel `a` has been used in the recorded run.
In a mutated run, GFuzz will select each of the cases with the same probability,
whether they have a potential partner or not. This means, that in many cases,
the mutation run will result in the timeout and therefore no new interesting
behavior will be triggered. By choosing cases with possible partners with a higher
probability, they will be chosen more often, resulting in fewer runs being
effected by timeout. This can reduce the number of runs needed to explore the
possible paths.

## Path expansion

> [!WARNING]
> This is not implemented yet

A program is interesting an will result in a new mutation, if it contains one of the following

- two `once.Do` on the same once that are concurrent
- a not successful `mutex.TryLock` or `mutex.TryRLock` that is concurrent with any other mutex lock on the same mutex
- two channel operations of the same type (both send or both receive) on the same channel

We can give an example for each of these, showing how they may help in finding potential bugs. Assume that the function `potentialBug()` contains a bug we want to find:

```go
func TestOnce(t *testing.T) {
	// If the Do 1 was executed during the first run, the code in the Do 2 function
	// was never executed. This means, we cannot find the potential bug.
	// By changing the oder of the two Do, we give the Do 2 the possibility
	// to execute the program code containing the potential bug
	var o sync.Once

	go func() {
		o.Do(func() {})  // Do 1
	}

	o.Do(potentialBug)   // Do 2
}
```


```go
func TestTryLock(t *testing.T) {
	// If the Lock was executed befoer the TryLock and the TryLock tried to execute
	// before the Unlock, the TryLock will fail, and we are not able to 
	// analyze the code in the TryLock Block. By switching the order of the TryLock
	// and Lock, we make it possible to find the potential bug
	m := sync.Mutex{}

	go func() {
		res := m.TryLock()
		if res {
			potentialBug()
			m.Unlock()
		}
	}

	m.Lock()
	// do something
	m.Unlock
}
```


```go
func TestChannel(t *testing.T) {
	// If the send 1 is executed before the send 2, the receive will get 
	// a value of a = 1 and will therefore not execute the if block with
	// the potential bug.
	// By changing the order of the two send operations, we can get the 
	// if block to be executed and are able to detect the potential bug.
	c := make(chan int, 2)

	go func() {
		c <- 1   // send 1
	}
	
	go func() {
		c <- 2   // send 2
	}

	if a := <-c; a == 2 {
		potentialBug()
	}
}
```


For each of the found instances a new mutation is created. 
To simplify the implementation and prevent the program from getting stuck during the recording of the mutations, the reordering is not done with the implemented replay mechanics, but with timers. 

To make sure, that (mostly) the same program is executed as before, the selects will chose there preferred case as implemented for the GFuzz mutations as the case that was executed in the program run the mutation is based on.

The mutations that are created here always aim at changing the order of two operations. To do this, the first operation of the two is marked in the `fuzzingData.log` file with there code position and a counter, showing which execution of this operation is targeted. Meaning if the send we want to delay at position `file:line` was executed 5 times and we want to delay the 3rd execution of this operation, we add `file:line:3` in the file `fuzzingData.log` file. When running the mutation, the implementation of the corresponding operations will check, if the fuzzingData contains such an operation based on the code position. If so, it will create a counter, starting with 1. If this counter is equal to the counter value in the fuzzingData, it will delay the execution of the operation by 2 seconds. If not, it will just increase the counter value. 

## fuzzingData
The file `fuzzingData.log` contains the data for a mutation to be executed. It contains of two blocks, separated by a separator
```log
[BlockSelect]
#
[BlockFlow]
```
The `BlockSelect` contains one line for each select in the mutation. Each line has the form 
```log
file:line;chosenIndex1,chosenIndex2,...
```
The `BlockFlow` contains a line for each operation that is delayed. Each line has the form
```
file:line;counterToDelayAt
```