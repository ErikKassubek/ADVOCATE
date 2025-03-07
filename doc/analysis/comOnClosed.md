## Communication on closed channel

## Send on closed"

Performing a send operation on a closed channel is a fatal operation.
We wish to identify send operations that could possible be performed on a closed channel.
For this purpose, we make use of `<HB`.
We see for a send event `e` on channel x where there is a close event `f` on channel x
such that neither `e <HB f` nor `f <HB e`.
The algorithm to carry out this check is as follows.


We keep track of the most recent send operations.
S(x) denotes the combined vector clock of all recent sends on channel x.
Initially, all entries in S(x) are set to zero.

S(x) is computed as follows.


~~~~~
sndRcvU(tS,tR,x) {
  V = sync(Th(tS), Th(tR))    -- Sync
  Th(tS) = V
  Th(tR) = V
  S(x) = sync(S(x), Th(tS))        --     record most recent send
  inc(Th(tS),tS)
  inc(Th(tR),tR)
}

snd(t,x,i) {
  X = X' ++ [(false,V)] ++ X''     -- S1
  Th(t) = sync(V, Th(t))           -- S2
  X = X' + [(true,Th(t))] + X''    -- S3
  S(x) = sync(S(x), Th(t))         --     record most recent send
  inc(Th(t),t)
  }
~~~~~~~~~


When processing `close` we check if there is any earlier send that could happen
concurrently to `close`.


~~~~
close(t,x) {
  Cl(x) = Th(t)
  if ! (S(x) < Cl(x)) {
     "send on closed"
  }
  inc(Th(t),t)
}
~~~~~~~~~

*The above is similar to checking if there is any earlier read that could conflict with a write.*


## "Receive on closed"

Performing a receive operation on a closed channel yields a default value.
Our tracing scheme records if such a case actually happened.

We wish to inform the user about possible receive operations that could be performed on a closed channel.
This is similar to "Send on closed", the difference is that we issue a warning message (as the user
might be aware that a receive on closed could happen).

The algorithm works in the same way as "Send on closed".
R(x) denotes the combined vector clock of all recent receives on channel x.
Initially, all entries in R(x) are set to zero.

R(x) is computed as follows.


~~~~~
sndRcvU(tS,tR,x) {
  V = sync(Th(tS), Th(tR))    -- Sync
  Th(tS) = V
  Th(tR) = V
  R(x) = sync(R(x), Th(tS))        --     record most recent receive
  inc(Th(tS),tS)
  inc(Th(tR),tR)
}

rcv(t,x,i) {
  X = [(true, i, V)] ++ X'            -- R1
  Th(t) = sync(V, Th(t))              -- R2
  X = X' ++ [(false, 0, Th(t))]       -- R3
  R(x) = sync(R(x), Th(t))            --     record most recent receive
  inc(Th(t),t)
}
~~~~~~~~~


When processing `close` we check if there is any earlier receive that could happen
concurrently to `close`.


~~~~
close(t,x) {
  Cl(x) = Th(t)
  if ! (R(x) < Cl(x)) {
     "receive on closed"
  }
  inc(Th(t),t)
}
~~~~~~~~~