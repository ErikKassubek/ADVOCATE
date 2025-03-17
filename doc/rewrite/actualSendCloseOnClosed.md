### Close on closed channel and actual send/recv on closed
We only record actually executed operations. For close on closed, we can therefore only detect actually occurring close on close. Reordering is therefore not necessary.
The same is true for actual send/recv on closed
