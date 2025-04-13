
# Trace rewrite

## Problem statement

Given a trace T and two events e and f.
Find a valid trace reordering T' of T where e and f appear right next to each other in the trace.

If such a T' exists, we say that for T', e and f are *next to each other reordered* wrt T,
written T' = nextTo(T,e,f).


## HB-based trace rewrite

Given a trace T where we have computed the <HB order for events in T.

Let e and f be two events in T where neither e <HB f nor f <HB e.
We show how to build a trace T' where T' = nextTo(T,e,f).

Suppose e appears before f in T

and

~~~~
T = T1 ++ [e] ++ T2 ++ [f] ++ T3
~~~~~~

where T1, T2 and T3 are subtraces and we write ++ for concatenation.

Take

~~~~
T' = T1 ++ T2' ++ [e,f]
~~~~~~~

where T2' = [ g | g in T2 and g <HB f ].


So, we push e "down" and T2 "up".
However, we only keep events in T2 that are in a HB relation with f.


Claim: T1 ++ T2' ++ [e,f] = nextTo(T1 ++ [e] ++ T2 ++ [f] ++ T3,e,f)
       where T2' = [ g | g in T2 and g <HB f ].


Note. Instead of [e,f] we can also consider [f,e] because e and f are not ordered under HB.

Proof.

Suppose g in T2 and e <HB g.
Then, g not in T2'.

Assume the contrary. If g in T2' we find that g <HB f.
In combination with the assumption e <HB g we derive e <HB f.
This contradicts the assumption that e and f are not ordered under HB.

Hence, we can argue that it is okay to put T2' "above" e.

Suppose g in T2 and g <MHB f.
That is, g must happen-before f where we assume that <MHB is derived from the [Go memory model](https://go.dev/ref/mem).
Our <HB relation conservatively approximates <MHB.
Hence, we conclude that g <HB f and therefore g in T2'.

Hence, we argume that T1 ++ T2' ++ [e,f] is a valid trace reordering.

QED.

### Case of thread-local traces

We rephrase "HB-based trace reconstruction" under the assumption that we have thread local traces.

* Events are ordered based on some total order <tr.

* The order <tr reflects the order in which events were processed when recording traces.

* Events are stored in thread-local traces.

* There are n threads where each thread i maintains its own trace L_i.


Let e in L_i and f in L_j be two events where neither e <HB f nor f <HB e.
We assume e <tr f.

Below is a description how to adjust thread-local traces such that e and f appear right
next to each other if we replay traces.


Consider event e's trace L_i.

~~~~
L_i = [ g | g in L_i and g <tr e ] ++ [e]
~~~~~~~~

We simply shorten the trace L_i by ignoring all events in thread i that were processed after e.


Consider event f's trace L_j.

~~~~~~~
L_j = [ g | g in L_i and g <tr e ]                             -- (1)
      ++
      [ g' | g' in L_i and e <tr g' and g' <tr f] ++ [f]       -- (2)
~~~~~~~~~~~~~~

Consider part (1). These are all events g that were processed before e. So, we keep them.

Consider part (2). These are all events g' in thread j that were processed after e and before f.
We have that g' <HB f by construction as f is part of thread j.

*Updating the global trace order*. When replaying the trace, events g's shall be processed
before event e. Hence, we need to build a new global trace order relation <trNew
such that g' <trNew e for each g'.
(This is mostly an implementation detail. Efficiency matters as traces may be large).

For all other threads k where k != i and k != j.

~~~~~~~~
L_k = [ g | g in L_k and g <tr e ]                                  -- (1)
      ++
      [ g' | in L_k and e <tr g' and g' <tr f and g' <HB f]         -- (2)
~~~~~~~~~~~

We again need to update the global trace order for events g's in part (2).


If we want to "flip" e and f we simply switch their (global) trace positions.


### General case of n events e1,...,en

We can generalize trace reconstruction in case there are n events e1,..,en
where for each ei and ej where i !=j we find that ei and ej are unordered under HB.

*Do we need the general case? It seems for the analysis scenarios we consider, consider two events is sufficient.
 For example, in case of deadlocks we could restrict ourselves to two threads involved.*



### What about non-atomic variables?

Currently, we do not observe any non-atomic variables.

~~~~~
Question: Does this affect the claim the reordering constructed above is valid?
~~~~~~~

Based on the Go memory model, non-atomic variables do not imply any (must) happens-before relations.
For example, consider the following program

~~~~{.go}
go func() {
  x = 1
}()


go func() {
  if(x >= 1) {
  ...
  }
}()

~~~~~~~~~

and a possible program run represented by the following trace

~~~~~~
    T1     T2

1.  wr(x)
2.         rd(x)
3.         ...
~~~~~~~~

There is a write-read dependency. So, it seems that "..." must happen after the write on x in T1.
However, the read and write are in a race. Racy programs imply undefined behavior.

~~~~~
Answer to the above question:

Assuming the program is race-free, we can argue that the reordering is valid.
Rigorously formalizing the statement might be quite a challenge though.
~~~~~~~~


## Reconstructions for the different analysis cases

- [send/recv on close](./rewrite/sendRecvOnClose.md)
- [actual send and close on closed](./rewrite/actualSendCloseOnClosed.md)
- [resource deadlocks](./rewrite/resourceDeadlock.md)
- [done before add](./rewrite/doneBeforeAddUnlockBeforeLock.md)
- [unlock before lock](./rewrite/doneBeforeAddUnlockBeforeLock.md)
- [leak](./rewrite/leaks.md)