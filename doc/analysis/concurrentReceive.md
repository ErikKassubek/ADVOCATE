### Analysis scenario: "Concurrent Receive"

Having multiple potentially concurrent receives on the same channel can cause
nondeterministic behavior, which is rarely desired. We therefor want to detect
such situations.

For this, we save the vector clock of the last receive for each combination of
channel and routine in L(R, x), where R is the routine, where the receive
took place and x the channel id. We now compare the current VC of R with all
Elements in L(R', x'), with R != R' and x == x'. If one of these vector clocks
is concurrent with our current VC, we have found a concurrent receive on the
same channel and will therefore return a warning.

Summarized that means, that the sendRcvU(tS, tR, x) and rcv(t, x, i) functions
are changed to

```
sendRcvU(tS, tR, x) {
  checkForConcurrentRecv(tR, x)
  ...
}

rcv(t, x, i) {
  checkForConcurrentRecv(t, x)
  ...
}
```

with

```
checkForConcurrentRecv(t, x) {   -- t: current routine, x: id of channel
  L(t, x) = Th(t)
  for routine, lastRecv in L {                -- Iterate over all routines
    if routine == t { continue }              -- Same routine
    if lastRecv[x] == nil { continue }        -- No receive for x in trace yet
    if !(Th(t) > lastRecv[x]) && !(Th(t) > lastRecv[x]) {  -- Is concurrent
      "concurrent recv"
    }
  }
}
```

This allows us to find concurrent receives on the same channel. It is not necessary to
search for concurrent send on the same channel, because this can behavior
can and often is useful, e.g. as a form of wait group.