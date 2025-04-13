### Leaks

#### Channel
There are two possible cases, either the stuck channel is unbuffered, or it is
buffered.

##### Unbuffered Channels
Buffered channels must communicate concurrently. If an unbuffered channel operation $c_s$/$c_r is
stuck, we check if there is a possible concurrent communication partner. If there
is non, we cannot rewrite the channel to get unstuck. This does not mean,
that it is impossible to run the program without getting stuck, e.g. with an
communication operation, that was not executed and therefore recorded in the run.
If there is, the operation must have communicated with another operation,
otherwise it would have communicated with $c$. Lets assume $s$ is the send
and $r$ the receive of this communication. We already know, that
$s$ must be concurrent with $r$, otherwise the communication wouldn't have
happened. We also know from the analysis, that $c_s$ and $r$ or $c_r$ and $s$
must be concurrent. Additionally we assume, that $c_s$ and $r$ or $c_r$ and
$s$ are also concurrent. If this is not the case, we assume, we cannot
rewrite the trace with the selected possible partner.
We can therefore rewrite the trace as:
~~~
T_1 ++ [X_s, c_s, r, X_e]
T_1 ++ [X_s, s, c_r, X_e]
~~~
In practice we will do this by deleting the original communication partner
and then removing all elements that happend after the stuck element or the
possible communication.
$X_s$ will only print a message, but not
effect the replay itself. $X_e$ will print a message and then disable the
replay. After this the program will continue to run, without following a
given trace.


##### Buffered Channels
Leaks on buffered channels do not depend on whether there is a concurrent
communication partner. There are two cases, in which a buffered channel
can refuse to send/receive and get stuck. Either the program tries to
send on a channel with a full channel buffer, or it tries to receive on
an empty buffer.

If a send $s$ is stuck, we check if there is are sends $s'$, that is concurrent
to $s$, but happened before it in the program run. If such $s'$ exists, we
reorder the trace in such a way, that $s$ is no before all those $s'$.\
Equivalently, is a recv $r$ is stuck, we try to find recvs, that are
concurrent to $r$, but happened before $r$ in the program run, and
order then, such that they happen after $r$. In practice, we remove them
and all other elements that are after them and in the same routine
from the trace such that $r$ is moved before them automatically, and let the
program run freely after executing $r$, by adding the $X_e$ control element.


#### Select
When searching for a possible communication partner, we check for all
cases, if there is a possible partner. When this partner is found, we
rewrite the program as if the select was only this selected channel operation.

#### Mutex
A mutex can only be blocked by a lock operation $l$. This operations blocks,
if the mutex is currently hold by another block operation. Because $l$
was blocked at the end of the program run, there is another lock operation $l'$,
which was fully executed, but there was no unlock operation. Because a potential later
unlock of this mutex was not recorded, it is not possible to try to move it
before $l$. Therefore we can only try to solve this stuck operation, if $l$ and
$l'$ are concurrent. In this case, we can try to execute $l$ before $l'$ to see,
if this prevents the stuck operation.
We therefore rewrite the trace from
~~~
T_1 ++ [l'] ++ T_2 ++ [l] ++ T_3
~~~~
to
~~~
T_1 ++ T_2' ++ [l, X_e]
~~~~
where X_e ends the guided replay and lets the rest of the program play out
by itself. T_2' is the set of all elements, that are before $l$.

#### Wait Group
Only the wait in a wait group can lead to an actual leak. This happens, when the
wait group counter is not 0 at any time after the wait command. We can
only influence the counter for the wait, by moving adds and dones, that are
concurrent to the wait. To minimize the counter as much as possible, we need
to move as many done before and as many add after the wait. We do this,
moving all elements, that are concurrent with the stuck wait to be after
the wait. To make sure, that we do not create a negative wait counter, we
keep the order of those moving elements the same as before the rewrite.

#### Conditional Variables
For a conditional variable only the wait operation can block. The block can be
ended by a Signal or Broadcast call.

If there is a Signal $s$, that is concurrent to the blocking wait $w$, there are
two possible cases. Either $s$ was executed before $w$ or $s$ was executed
after $w$ but released another wait $w'$.\
If $w'$ is HB before $w$, we cannot use this signal to release $w$, because
signal wakes the wait in the order in which the waits where started. If
$w'$ is concurrent with $w$, we move $w$ to be before $w'$. If $w$ was executed
after $s$, we move $s$ to be before $w$. Unfortunately, waits are always
surrounded by lock operations, which in out HB relation scheme create a happens
before relation. With this, this type of reorder is not possible. For this
reason, it is currently necessary to run the analyzer with `-c`, which disables
the happens before relation of mutex operations.

If there is a Broadcast call $b$, that is concurrent to the blocking wait $w$,
but happened before it in the program run, we can move $b$ to
be after $w$, by moving all elements that are concurrent to $b$ to be before $b$.


