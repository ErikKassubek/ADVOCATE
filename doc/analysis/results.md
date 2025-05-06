# Analysis Result

The found problems found during the analysis are stored in two different formats.

The first format is a machine readable format, which is stored in the file `results_machine.log`.
It is used to further process the results, mainly for the rewriting and replaying of the trace.

The second format is a human readable format, which is stored in the file `results_readable.log`
and printed to the terminal. It is used to show the results to the user.

## Machine readable result file

The result file contains all potential bugs found in the analyzed trace.

The file contains one line for each found problem. The line has the following form
```
[typeID],[args]
```
or
```
[typeId],[args],[args]
```
with
```
[args]: [arg] | [arg];[arg] | [arg];[arg];[arg] | ...
[arg] : T:[routineId]:[objId]:[tPre]:[objType]:[file]:[line] (trace element)
[arg] : S:[objId]:[objType] (select case)
```
The typeIDs have the following meaning:

- A00: Unknown panic
- A01: Send on closed channel
- A02: Receive on closed channel
- A03: Close on closed channel
- A04: Close on nil channel
- A05: Negative wait group
- A06: Unlock of not locked mutex
- A07: Concurrent recv
- P01: Possible send on closed channel
- P02: Possible receive on closed channel
- P03: Possible negative waitgroup counter
- P04: Possible unlock deadlock before lock
- P05: Possible cyclic deadlock
- L01: Leak on unbuffered channel with possible partner
- L02: Leak on unbuffered channel without possible partner
- L03: Leak on buffered channel with possible partner
- L04: Leak on buffered channel without possible partner
- L05: Leak on nil channel
- L06: Leak on select with possible partner
- L07: Leak on select without possible partner (includes nil channels)
- L08: Leak on mutex
- L09: Leak on waitgroup
- L10: Leak on cond
- R01: Unknown panic in recording
- R02: Timeout in recording

<!--P06: Possible mixed deadlock, disabled-->
`[args]` shows the elements involved in the problem. There are either
one or two, while the args them self can contain multiple trace elements or select cases.\
The arg in args are separated by a semicolon (;).\
Each arg contains the following elements separated by a colon (:)
- `[routineId]` is the id of the routine that contains the operation
- `[objId]` is the id of the object that is involved in the operation
- `[tPre]` is the time of the operation
- `[opjType]` is the type of the element
	- Atomic:
	  - AL: Load
		- AS: Store
		- AA: Add
		- AW: Swap
		- AC: CompSwap
  - Channel:
    - CS: Send
    - CR: Receive
    - CC: Close
  - Mutex:
    - ML: Lock
    - MR: RLock
    - MT: TryLock
    - MY: TryRLock
    - MU: Unlock
    - MN: RUnlock
  - Waitgroup:
    - WA: Add
    - WD: Done
    - WW: Wait
  - Select:
    - SS: Select
  - Cond:
    - DW: Wait
    - DB: Broadcast
    - DS: Signal
  - Once:
    - OE: Done Executed
    - ON: Done Not Executed (because the once was already executed)
  - Routine:
    - RF: Fork
		- RE: End
	- New:
		- NA: new atomic variable (not used)
		- NC: new channel
		- ND: new cond (not used)
		- NM: new mutex (not used)
		- NO: new once (not used)
		- NW: new waitGroup (not used)
- `[file]` is the file of the operation in the program code
- `[line]` is the line of the operation in the program code

## Human readable result file

The result file contains all potential bugs found in the analyzed trace.
A possible result would be:
```
Possible send on closed channel:
	close: example.go:10@47
	send: example.go:40@44
Possible receive on closed channel:
	close: example.go:10@47
	recv: example.go:20@43
Possible negative waitgroup counter:
	add: example.go:50@77;
	done: example.go:60@80;
```
Each found problem consist of three lines (the third line can be empty).
The first line explains the
type of the found bug. The other two line contain the information about the
elements responsible for the problem. The elements always have the
form of
```
[type]: [file]:[line]@[tPre]
```


## Results

The following examples show the different types of problems that can be found during the analysis.

### Send on close
A send on closed is an actual send on a closed channel (always leads to panic).
The two args of this case are:

- the send operation
- the close operation

An example for a send on closed is:
```golang
1 func main() {          // routine = 1
2 	c := make(chan int)  // objId = 2
3 	close(c)             // tPre = 10
4 	c <- 1               // tPre = 12
5 }
```


In the machine readable format, the send on closed has the following form:
```
A01,T:1:2:12:CS:example.go:4,T:1:2:10:CC:example.go:3
```
In the human readable format, the send on closed has the following form:
```
Found receive on closed channel:
	send: example.go:4@12
	close: example.go:5@10
```


### Receive on close
A receive on closed is an actual receive on a closed channel.
The two args of this case are:

- the receive operation
- the close operation

An example for a receive on closed is:
```golang
1 func main() {          // routine = 1
2   c := make(chan int)  // objId = 2
3   close(c)             // tPre = 10
4   <-c                  // tPre = 12
5 }
```

In the machine readable format, the receive on closed has the following form:
```
A02,T:1:2:12:CR:example.go:4,T:1:2:10:CC:example.go:3
```

In the human readable format, the receive on closed has the following form:
```
Found receive on closed channel:
	recv: example.go:4@12
	close: example.go:5@10
```

### Close on close
A send on closed is an actual close on a closed channel (always leads to panic).
The two args of this case are:

- the close operation that leads to the panic
- the close operation that is the first close on the channel

An example for a close on closed is:
```golang
1 func main() {          // routine = 1
2   c := make(chan int)  // objId = 2
3   close(c)             // tPre = 10
4   close(c)             // tPre = 12
5 }
```


In the machine readable format, the close on closed has the following form:
```
A03,T:1:2:12:CC:example.go:4,T:1:2:10:CC:example.go:3
```

In the human readable format, the close on closed has the following form:
```
Found close on closed channel:
	close: example.go:4@12
	close: example.go:3@10
```

### Close on nil
A close on nil is an actual close on a nil channel (always leads to panic).
The two args of this case are:

- the close operation that leads to the panic

An example for a close on closed is:
```golang
1 func main() {          // routine = 1
2   c := make(chan int)  // objId = 2
3   c = nil
4   close(c)             // tPre = 12
5 }
```


In the machine readable format, the close on closed has the following form:
```
A04,T:1:2:12:CC:example.go:4
```

In the human readable format, the close on closed has the following form:
```
Found close on nil channel:
	close: example.go:4@12
```

### Negative wait group
A negative wait group is an actual negative wait group counter.
The one argument of this case is:

- the done that resulted in the negative wg counter

An example for a close on closed is:
```golang
1 func main() {          // routine = 1
2   var wg sync.WaitGroup  // objId = 2
3   wg.Add(1)              // tPre = 10
4   wg.Done()              // tPre = 12
5   wg.Done()              // tPre = 14
6 }
```


In the machine readable format, the close on closed has the following form:
```
A05,T:1:2:14:WD:example.go:5
```

In the human readable format, the close on closed has the following form:
```
Found actual negative wait group counter:
	done: example.go:5@14

```

### Unlock of not locked mutex
A unlock on a not locked mutex was detected
The argument in this case is

- the unlock operation

An example is
```golang
1 func main() {        // routine = 1
2   var m sync.Mutex   // objId = 2
3   m.Lock()           // tPre = 10
4   m.Unlock()         // tPre = 12
5   m.Unlock()         // tPre = 14
}
```
In the machine readable format, the close on closed has the following form:
```
A06,T:1:2:14:MU:example.go:5
```

In the human readable format, the close on closed has the following form:
```
Found unlock on not locked mutex:
	unlock: example.go:5@14

```


### Concurrent recv
A concurrent recv shows two receive operations on the same channel that are concurrent.:
The two args of this case are:

- the recv operation
- the recv operation

An example for a concurrent recv is:
```golang
 1 func main() {           // routine = 1
 2   c := make(chan int)   // objId = 2
 3
 4   go func() {           // routine = 2
 5     <-c                 // tPre = 10
 6   }()
 7
 8   go func() {           // routine = 3
 9     <-c                 // tPre = 20
10   }()
11
12   c <- 1                // tPre = 30
10 }
```

The machine readable format of the concurrent recv has the following form:
```
A07,T:3:2:20:CR:example.go:9,T:4:2:10:CR:example.go:5
```
The human readable format of the concurrent recv has the following form:
```
Found concurrent Recv on same channel:
	recv: example.go:9@20
	recv: example.go:5@10
```

### Possible send on closed

A possible send on closed is a possible but not actual send on a closed channel.
The two args of this case are:

- the send operation
- the close operation

An example for a possible send on closed is:

```golang
1 func main() {          // routine = 1
2   c := make(chan int)  // objId = 2
3
4   go func() {          // routine = 2
5     c <- 1             // tPre = 10
6   }()
7
8   go func() {         // routine = 3
9     <- c              // tPre = 20
10  }()
11
12  close(c)            // tPre = 30
12 }
```


In the machine readable format, the possible send on closed has the following form:

```
P01,T:2:2:10:CS:example.go:5,T:3:2:30:CC:example.go:30
```

```
Possible send on closed channel::
	send: example.go:5@10
	close: example.go:12@30
```

### Possible recv on closed
A possible recv on closed is a possible but not actual recv on a closed channel.
The two args of this case are:

- the recv operation
- the close operation

An example for a possible recv on closed is:

```golang
1 func main() {          // routine = 1
2   c := make(chan int)  // objId = 2
3
4   go func() {          // routine = 2
5     c <- 1             // tPre = 10
6   }()
7
8   go func() {         // routine = 3
9     <- c              // tPre = 20
10  }()
11
12  close(c)            // tPre = 30
12 }
```


In the machine readable format, the possible send on closed has the following form:

```
P02,T:3:2:20:CR:example.go:9,T:3:2:30:CC:example.go:30
```

```
Possible send on closed channel:
	recv: example.go:9@20
	close: example.go:12@30
```


### Possible negative waitgroup counter

A possible negative waitgroup counter is a possible but not actual negative waitgroup counter.
The two args of this case are:
- The list of add operations that might make the counter negative (separated by semicolon)
- The list of done operations that might stop the counter from become negative (separated by semicolon)

An example for a possible negative waitgroup counter is:
```golang
 1 func main() {            // routine = 1
 2   var wg sync.WaitGroup  // objId = 2
 3
 4   go func() {            // routine = 2
 5     wg.Add(1)            // tPre = 10
 6   }()
 7   go func() {            // routine = 3
 8     wg.Add(1)            // tPre = 20
 9   }()
10
 8   go func() {            // routine = 4
 9     wg.Done()            // tPre = 30
10  }()
11
12   wg.Done()            // tPre = 40
13 }
```

The machine readable format of the possible negative waitgroup counter has the following form:
```
P03,T:2:2:10:WA:example.go:5;T:3:2:20:WA:example.go:8,T:4:2:30:WD:example.go:9;T:1:2:40:WD:example.go:12
```

The human readable format of the possible negative waitgroup counter has the following form:
```
Possible negative waitgroup counter:
	add: example.go:5@10;example.go:8@20
	done: example.go:9@30;example.go:12@40
```

### Leak on unbuffered channel
#### With possible partner
A leak on an unbuffered channel with a possible partner is a unbuffered channel is leaking,
but has a possible partner.
The two arg of this case is:

- the channel that is leaking
- the possible partner of the channel

An example for a leak on an unbuffered channel with a possible partner is:
```golang
1 func main() {          // routine = 1
2   c := make(chan int)  // objId = 2
3
4   go func() {          // routine = 2
5     c <- 1             // tPre = 10
6   }()
7
8   go func() {          // routine = 3
9     <- c               // tPre = 20
10  }()
11
12  go func() {          // routine = 4
13    c <- 1             // tPre = 30
14  }()
15 }
```
We assume that line 5 send to line 9 and line 13 is leaking.

The machine readable format of the leak on an unbuffered channel with a possible partner has the following form:
```
L01,T:4:2:30:CS:example.go:13,T:3:2:20:CR:example.go:9
```

The human readable format of the leak on an unbuffered channel with a possible partner has the following form:
```
Leak on unbuffered channel with possible partner:
	channel: example.go:13@30
	partner: example.go:9@20
```


#### Without possible partner
A leak on an unbuffered channel without a possible partner is a unbuffered channel that is leaking,
but has no possible partner.

The one arg of this case is:

- the channel that is leaking

An example for a leak on an unbuffered channel without a possible partner is:
```golang
1 func main() {          // routine = 1
2   c := make(chan int)  // objId = 2
3
4   go func() {          // routine = 2
5     <-c                // tPre = 10
6   }()
7 }
```

The machine readable format of the leak on an unbuffered channel without a possible partner has the following form:
```
L02,T:2:2:10:CR:example.go:5
```

The human readable format of the leak on an unbuffered channel without a possible partner has the following form:
```
Leak on unbuffered channel without possible partner:
	channel: example.go:5@10

```

### Leak on buffered channel
#### With possible partner
A leak on an buffered channel with a possible partner is a buffered channel is leaking,
but has a possible partner.
The two arg of this case are:

- the channel that is leaking
- the possible partner of the channel

An example for a leak on an buffered channel with a possible partner is:
```golang
1 func main() {             // routine = 1
2   c := make(chan int, 1)  // objId = 2
3
4   go func() {             // routine = 2
5     c <- 1                // tPre = 10
6   }()
7
8   go func() {             // routine = 3
9     <- c                  // tPre = 20
10  }()
11
12  go func() {             // routine = 4
13    c <- 1                // tPre = 30
14  }()
15 }
```
We assume that line 5 send to line 9 and line 13 is leaking.

The machine readable format of the leak on an buffered channel with a possible partner has the following form:
```
L03,T:4:2:30:CS:example.go:13,T:3:2:20:CR:example.go:9
```

The human readable format of the leak on an buffered channel with a possible partner has the following form:
```
Leak on buffered channel with possible partner:
	channel: example.go:13@30
	partner: example.go:9@20
```


#### Without possible partner

A leak on an buffered channel without a possible partner is a buffered channel that is leaking,
but has no possible partner.

The one arg of this case is:

- the channel that is leaking

An example for a leak on an unbuffered channel without a possible partner is:
```golang
1 func main() {              // routine = 1
2   c := make(chan int, 1)   // objId = 2
3
4   go func() {              // routine = 2
5     <-c                    // tPre = 10
6   }()
7 }
```

The machine readable format of the leak on an unbuffered channel without a possible partner has the following form:
```
L04,T:2:2:10:CR:example.go:5
```

The human readable format of the leak on an unbuffered channel without a possible partner has the following form:
```
Leak on buffered channel without possible partner:
	channel: example.go:5@10

```

### Leak on nil channel

A leak on a nil channel is a nil channel trying to communicate.

The one arg of this case is:

- the nil channel that is leaking

An example for a leak on a nil channel is:
```golang
func main() {
	var c chan int
	go func() {
		c <- 1
	}()
}
```

The machine readable format of the leak on a nil channel has the following form:
```
L05,T:2:-1:10:CS:example.go:6
```

The human readable format of the leak on a nil channel has the following form:
```
Leak on nil channel:
	channel: example.go:6@10

```

### Leak on select
#### With possible partner
A leak on an select with a possible partner is a select is leaking,
but has a possible partner.
The two arg of this case is:

- the select that is leaking
- the possible partner of the channel or select

An example for a leak on an select with a possible partner is:
```golang
 1 func main() {             // routine = 1
 2   c := make(chan int)     // objId = 2
 3
 4   go func() {             // routine = 2
 5     c <- 1                // tPre = 10
 6   }()
 7
 8   go func() {             // routine = 3
 9     <- c                  // tPre = 20
10   }()
11
12   go func() {             // routine = 4
13     select {              // objId = 3, tPre = 30
14     case c <- 1:
15     }
14   }()
15 }
```
We assume that line 5 send to line 9 and that the select is leaking.

The machine readable format of the leak on an select with a possible partner has the following form:
```
L06,T:4:3:30:SS:example.go:13,T:3:2:20:CR:example.go:9
```

The human readable format of the leak on an select with a possible partner has the following form:
```
Leak on select with possible partner:
	select: example.go:13@30
	partner: example.go:9@20
```

#### Without possible partner
A leak on an select without a possible partner is a select that is leaking,
but has no  partner.
The one arg of this case is:

- the select that is leaking

```golang
1 func main() {             // routine = 1
2   c := make(chan int)     // objId = 2
3
4   go func() {             // routine = 2
5     select {              // objId = 3, tPre = 10
6     case c <- 1:
7     }
8   }()
9  }
```

The machine readable format of the leak on an select without a possible partner has the following form:
```
L07,T:2:3:10:SS:example.go:5
```

The human readable format of the leak on an select without a possible partner has the following form:
```
Leak on select without possible partner:
	select: example.go:5@10

```

### Leak on mutex

A leak on a mutex is a mutex that is leaking.
The two arg of this case is:

- the mutex operation that is leaking
- the last lock operation on the mutex

An example for a leak on a mutex is:
```golang
1 func main() {          // routine = 1
2   var m sync.Mutex     // objId = 2
3
4   go func() {          // routine = 2
5     m.Lock()           // tPre = 20
6   }()
7
8  m.Lock()              // tPre = 10
9 }
```
We assume, that the lock operation in line 10 happened before the lock operation in line 5.
The Lock operation in line 5 is leaking.

The machine readable format of the leak on a mutex has the following form:
```
L08,T:2:2:20:ML:example.go:5,T:1:2:10:ML:example.go:8
```

The human readable format of the leak on a mutex has the following form:
```
Leak on mutex:
	mutex: example.go:5@20
	last: example.go:8@10
```

### Leak on waitgroup

A leak on a waitgroup is a waitgroup that is leaking.
The one arg of this case is:

- the waitgroup operation (wait) that is leaking

An example for a leak on a waitgroup is:
```golang
1 func main() {          // routine = 1
2   var wg sync.WaitGroup // objId = 2
3
4   wg.Add(1)          // tPre = 20
5
6   wg.Wait()            // tPre = 10
7 }
```

The machine readable format of the leak on a waitgroup has the following form:
```
L09,T:1:2:10:WW:example.go:6
```

The human readable format of the leak on a waitgroup has the following form:
```
Leak on waitgroup:
	waitgroup: example.go:6@10

```

### Leak on cond

A leak on a cond is a cond that is leaking.
The one arg of this case is:

- the cond operation (wait) that is leaking\

An example for a leak on a cond is:
```golang
1 func main() {          // routine = 1
2   var c sync.Cond      // objId = 2
3
4   c.Wait()             // tPre = 20
5 }
```

The machine readable format of the leak on a cond has the following form:
```
L10,T:1:2:20:NW:example.go:4
```

The human readable format of the leak on a cond has the following form:
```
Leak on cond:
	cond: example.go:4@20

```
