### Done before Add and Unlock before Lock

The analysis for done before add and unlock before lock are basically identical.
For this reason, the rewrite is also identical. The following is an overview for
the done before add case. For the unlock before lock case, just replace every
wait group with mutex, add with lock and done with unlock.

We assume, that the program run that was analyzed did not result in a negative
wait counter. Therefore we know, that for a wait group, the number of add is greater or
equal than the number of done. From the analysis, we get an incomplete but
optimal bipartite matching between adds and dones. For all dones $D$, that are
not part of this matching, meaning they do not have an earlier add associated
with it, we can find a unique add $A$, that is concurrent to $D$. For
each of those $D$ we now shift all elements that are concurrent or after $D$
to be after that $D$. This includes the $A$.