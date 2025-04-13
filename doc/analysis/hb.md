# Happens before

The analysis tries to calculate happens-before relationships between
the different recorded operations using [Vector Clocks](https://en.wikipedia.org/wiki/Vector_clock).

The following describes how to derive happens-before relations and build vector clocks for Go synchronization primitives.
To remain efficient, the happens-before relations we derive over-approximate the *must* happens-before relations.
That is, we might rule out certain reorderings. This then leads to false negatives.

To derive the happens-before relations,
we use as a reference [The Go Memory Model](https://go.dev/ref/mem).
We will write `<HB` to denote the happens-before ordering relation.

Our tracing scheme records events in thread-local traces.
Traces are lists with the head element being the first recorded event in that trace and so on.
We use a counter to represent the global time.
Each event is annotated with a pre/post counter.
The pre counter represents the time before the underlying operation is executed.
The post counter represents the time after the operation could be successfully executed.

The following conditions are guaranteed by our tracing scheme:


* post >= pre + 1

* send_post > reiceive_post where send and receive communicate with each other via a channel

* if post = 0 the operation is incomplete (for example blocked receive)

* For (thread-local) trace T and T = [..., e, ..., f, ...] we have that e_post < f_post

For the purpose of representing thread-local traces as a single trace that represents the interleaved execution of the program
when can use the post counter to identify the 'trace position' of an event.


## Vector clocks


We view a vector clock as a list of time stamps
where the time stamp's position correspond to a thread id.

Vector clock operations.

~~~~
Increment the time stamp of thread i

    inc([k1,...,ki-1,ki,ki+1,...,kn],i) = [k1,...,ki-1,ki+1,ki+1,...,kn]

Synchronize two vector clocks by taking the larger time stamp

     sync([i1,...,in],[j1,...,jn]) = [max(i1,j1), ..., max(in,jn)]
~~~~~~~~~


In some implementation, it will be the easiert to use a map (ThreadID to TimeStamp) to represent vector clocks.

The time stamp of the main thread is one wherer all other entries are set to zero.

## Fork (spawn)

Events:

~~~
fork(t1,t2)      in thread t1 we start the new thread t2
~~~~~

~~~~
fork(t1,t2) {
  Th(t2) = Th(t1)
  inc(Th(t2),t2)
  inc(Th(t1),t1)
}
~~~~~~~~~


## (RW)-Mutex

RW Mutex = multiple readers but only single writers

Events:

~~~~
rlock(t,x)           -- thread t executes rlock of x
runlock(t,x)
lock(t,x)
unlock(t,x)
~~~~~~~~~~

Go also supports trylock/tryrlock.
If successful this corresponds to lock/rlock, otherwise trylock/tryrlock can be ignored.


Based on the Go memory model for [RWMutex](https://go.dev/ref/mem#locks), happens-before relations are:

~~~

(RW-1) unlock(_,x)_i <HB lock(_,x)_j where i < j

(RW-2) runlock(_,x)_i <HB lock(_,x)_j where i < j

(R#-3) unlock(_,x)_i <HB rlock(_,x)_j where i < j


where i and j denote the trace position.
~~~~~~


For each RWMutex we introduce vector clocks RelW(x) and RelR(x)
to record the release 'time' (vector clock) of a lock.
Initially, all entries in RelW(x) and RelR(x) are equal to zero.
We assume that Th(t) holds the vector clock of thread t.


Event processing functions to compute vector clocks are as follows.


~~~~~
lock(t,x) {
  Th(t) = sync(Th(t), RelW(x))     -- RW-1
  Th(t) = sync(Th(t), RelR(x))     -- RW-2
  inc(Th(t),t)
}

unlock(t,x) {
  RelW(x) = Th(t)       -- RW-1
  RelR(x) = Th(t)       -- RW-2
  inc(Th(t),t)
}


rlock(t,x) {
  T(t) = sync(Th(t), RelW(x))     -- RW-3
  inc(Th(t),t)
}

runlock(t,x) {
  RelR(x) = sync(RelR(x),Th(t))
            -- rlock and runlock do not synchronize.
      -- lock synchronizes with any prior runlock, see RW-2
      -- Hence, when lock synchronizes with RelR(x),
      -- RelR(x) must represent *all* prior runlocks.
      -- This is achieved by merging the vector clocks of all prior runlocks.
  inc(Th(t),t)
}
~~~~~~~~~~~


Example

~~~~
      #1        #2        #3

1.  rlock(y)
2.              rlock(y)
3.  runlock(y)
4.              runlock(y)
5.                          lock(y)
6.                          unlock(y)
~~~~~~~~~~~~

Note. There is a valid reordering under which the events in thread #3 are executed before the other events.
As we over-approximate, the happens-before relation orders critical sections based on their "textual" order
in the trace. Hence, we find that `runlock(y)_3 <HB lock(y)_5` and `runlock(y)_4 <HB lock(y)_5`.

This ordering based on "textual" order can be turned of. In this case all operations are replaced by

~~~~
op(t, x) {
  inc(Th(t), t)
}
~~~~

This can lead to more bugs being detected, but can also introduce false positives.

## Atomic

Events refer to memory locations.
Our tracer supports the following types of operations.
They can be mapped to reads and writes.

* Load = read
* Store = write
* Add = write
* Swap = read; write
* CompareAndSwap = read; write

Based on the [Go Memory Model](https://go.dev/ref/mem#atomic),
the behavior of atomic variables corresponds to Java's volatile variable.
So, a read synchronizes with the most recent write.

LW(x) records the vector clock of the last (atomic) write of x.
Initially, all entries in LW(x) are set to zero.

Event processing functions are as follows.

~~~~
write(t,x) {
  LW(x) = Th(t)
  inc(Th(t),t)
}

read(t,x) {
  Th(t) = sync(LW(x),Th(t))
  inc(Th(t),t)
}


~~~~~~~~~

## Wait groups

Events:

~~~~
add(t,g)
done(t,g)
wait(t,g)
~~~~~

Based on the description of [wait groups](https://pkg.go.dev/sync#WaitGroup), happens-before relations are:


~~~~
done(_,g)_i <HB wait(_,g)_j where i < j
~~~~~~~~


The wait group description says:

* Calls with a negative delta, or calls with a positive delta that start when the counter is greater than zero, may happen at any time.

*Does this mean that add acts like done? Yes!*


Our tracer does not explicitly distinguish between `add` and `done`.
So, it's okay to treat both the same way.

~~~~
add(_,g)_i <HB wait(_,g)_j where i < j
~~~~~~~~

Further assumptions are:

* A1: The initial `add` must take place before any `done` and `wait`.

* A2: If the wait group `g` is used again, any resetting via `add` must happen after any prior `wait`.



This guarantees that there is no need to "reset" WG(g).


Event processing is as follows.
For each wait group g, we assume a vector clock WG(g).
Initially, all entries in WG(g) are set to zero.

~~~~
add(t,g) {
  WG(g) = sync(WG(g),Th(t))
  inc(Th(t),t)
}

done(t,g) {
  WG(g) = sync(WG(g),Th(t))
  inc(Th(t),t)
}

wait(t,g) {
  Th(t) = sync(WG(g),Th(t))
  inc(Th(t),t)
}
~~~~~~~~


## Channels


A send communicating with a receive is identified via a unique id i.


Events:

~~~~
snd(t,x,i)      -- unbuffered send on x with communication partner i in thread t
rcv(t,x,i)
sndB(t,x,i)     -- buffered version, we assume that size(x) denotes the buffer size
rcvB(t,x,i)

close(t,x)      -- closing a channel
rcvC(t,x)       -- receive on a closed channel (send on closed fails immediately)
~~~~~~~


When processing events, we use the post counter to identify the next to be processed events.

### Unbuffered

The next to be processed event is some unbuffered send `send(tS,x,k)` in some thread-local trace T_i.
Via the communication id `k` via can find its communication partner `rec(tR,x,k)`.
By construction, `rec(tR,x,k)` must be the top element of some thread-local trace T_j.


T_i = [send(tS,x,k), ...]

T_j = [rec(tR,x,k), ...]


We drop send(tS,x,k) and rec(tR,x,k) from the trace and carry out the following call.


~~~~
sndRcvU(tS,tR,x) {
  V = sync(Th(tS), Th(tR))    -- Sync
  Th(tS) = V
  Th(tR) = V
  inc(Th(tS),tS)
  inc(Th(tR),tR)
}
~~~~~~~~~

Vector clocks of sender and receiver are synchronized. See `Sync`.

### Buffered

More interesting is the computation (of the happens-before relations via vector clocks)
for the case of buffered channels.
Based on the [Go memory model for channels](https://go.dev/ref/mem#chan)
we find the following requirements:


* REQ-CHAN-1: A send on a channel happens before the corresponding receive from that channel completes.

* REQ-CHAN-2: The kth receive on a channel with capacity C happens before the k+Cth send from that channel completes.


Based on this requirement, we compute vector clocks as follows.


For each buffered channel x of size n, we assume a list X of size n.
Elements in X are tuples of type (bool, int, VectorClock).

Initially,

X = [(false, 0, V0),...,(false, 0, V0)]

where V0 is the vector clock where all entries are set to zero
and false indicates that the buffer space is empty.
The `int` paramter is necessary to identfy communication partners.

We write

X = X' ++ [(false, 0, V)] ++ X''


to denote that elements in X' are all occupied.


Event processing for snd/rcv is based on the total order as specified by the post counter.


~~~~~
snd(t,x,i) {
  X = X' ++ [(false, 0, V)] ++ X''    -- S1
  Th(t) = sync(V, Th(t))              -- S2
  X = X' + [(true, i, Th(t))] + X''   -- S3
  inc(Th(t),t)
}

rcv(t,x,i) {
  X = [(true, i, V)] ++ X'            -- R1
  Th(t) = sync(V, Th(t))              -- R2
  X = X' ++ [(false, 0, Th(t))]       -- R3
  inc(Th(t),t)
}
~~~~~~~~~~~

* Send puts its vector clock and communication id i in the next available buffer slot. See `S3`.

* A subsequent receive fetches this vector clock to synchronize. The communication id of the receive must match the communication in the buffer slot. See `R1` and `R2`.

This guarantees REQ-CHAN-1.

Additionally, we perform the following.

* The receiver puts its vector clock in the now freed buffer slot. See `R3`.

* If a subsequent sender useses this buffer slut, the sender synchronizes with the receiver. See `S1` and `S2`.

This guarantees REQ-CHAN-2.

*Somewhere in the Go spec it says that buffered channels behave like FIFO channels. But the above rules do seem not capture the FIFO property!*

#### Enforce FIFO Channels (optional)

*The following ppplies to buffered channels only*.

Based on the total order among sends/receives as found in the trace, we can enforce the FIFO property.

LastSnd(x) records the vector clock of the last send on channel x.
Initially, all entries in LastSnd(x) are set to zero.

Processing of `snd(t,x,i)` is adapted as follows.

~~~~~
snd(t,x,i) {
  X = X' ++ [(false,V)] ++ X''     -- S1
  Th(t) = sync(V, Th(t))           -- S2
  Th(t) = sync(LastSnd(x), Th(t))  -- S2'
  LastSnd(x) = Th(t)               -- S2''
  X = X' + [(true,Th(t))] + X''    -- S3
  inc(Th(t),t)
}
~~~~~~~~~

Similarly, we can adapt the (buffered) receive case.

### Closed and receive on closed

The Go memory model specifies:


*The closing of a channel is synchronized before a receive that returns a zero value because the channel is closed. *

The resulting vector clock computations are as follows.


Cl(x) records the vector clock of the close operation on channel x.

~~~~
close(t,x) {
  Cl(x) = Th(t)
  inc(Th(t),t)
}


rcvC(t,x) {
  Th(t) = sync(Cl(x), Th(t))
  inc(Th(t),t)
}

~~~~~~~~


## Once

Events:

~~~~
onceT(t,x)         -- successful once x
onceF(t,x)         -- failed once x
~~~~~~

The assumption is:

* onceT(t,x)_post < onceF(t,x)<post


[Once description](https://go.dev/ref/mem#once):


* Multiple threads can execute once.Do(f) for a particular f, but only one will run f(), and the other calls block until f() has returned.

* The completion of a single call of f() from once.Do(f) is synchronized before the return of any call of once.Do(f).

*Function call f() is executed as part of the "winning" thread*.


In terms of the happens-before we find the following.

~~~~~
   onceT(_,x) <HB onceF(_,x)
~~~~~~~~~


O(x) records the vector clock of `onceT(_,x)`.
Initially, all entries in O(x) are zero.


~~~~~
onceT(t,x) {
   O(x) = Th(t)
   inc(Th(t),t)
}

onceF(t,x) {
   Th(t) = sync(O(x), Th(t))
   inc(Th(t),t)
}
~~~~~~~~



## Condition variables

Events:

~~~
Wait(t, x)          -- Wait
Signal(t, x)        -- Release on wait
Broadcast(t, x)     -- Release all
~~~

[Conditional variables description](https://pkg.go.dev/sync#Cond)

* Broadcast: Broadcast wakes all goroutines waiting on x.
* Signal: Signal wakes one goroutine waiting on x, if there is any. The routine that is woken is decided by a ticketing system (longest waiting).
* Wait: Wait atomically unlocks c.L and suspends execution of the calling goroutine.

In terms of the happens-before we find the following.

~~~
   signal(_,x) <HB wait(_,x) [0]  -- first not considered wait
   broadcast(_,x) <HB wait(_,x)   -- all not considered wait
~~~

For each conditional variable we save the currently waiting routines in Cond(x).
The update of the vector clocks is implemented as follows:

~~~~
condWait(t, x) {
  Cond(x).append(t)
  inc(Th(t), t)
}

condSignal(t, x) {
  if len(Cond(x)) != 0 {
    tWait = cond(x).Pop(0)  -- remove and return the first element
    sync(Th(tWait), Th(t))
  }
    inc(Th(t), t)
}

condBroadcast(t, x) {
  for tWait in Cond(x) {
    sync(Th(tWait), Th(t))
  }
  Cond(x).Clear()  -- remove all element from Cond(x)
  inc(Th(t), t)
}

~~~~

## Examples

Consider the trace

~~~~~~
   T1                  T2
1. fork(T2)
2.                     snd(c,1)
3. rcv(c,1)
4. close(c)
~~~~~~~~~~


We annotate the trace with vector clock information.



~~~~~~
   T1                  T2           S(c)
   [1,0]
1. fork(T2)
                       [1,1]
   [2,0]
2.                     snd(c,1)
3. rcv(c,1)

    call sndRcvU(T1,T2,c):

    sync([2,0],[1,1]) = [2,1]
                                     [2,1]


    [3,1]
                       [2,2]
4. close(c)

   [2,1] < [3,1] => "okay"

   [4,1]
~~~~~~~~~~

