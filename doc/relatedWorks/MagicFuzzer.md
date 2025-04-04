# MagicFuzzer: Scalable Deadlock Detection for Large-Scale Applications

[Y. Cai and W. K. Chan, "MagicFuzzer: Scalable deadlock detection for large-scale applications," 2012 34th International Conference on Software Engineering (ICSE), Zurich, Switzerland, 2012, pp. 606-616, doi: 10.1109/ICSE.2012.6227156.](https://ieeexplore.ieee.org/document/6227156)

## Summary
The Goal of MagicFuzzer is to detect and conform the existence of cyclic mutex deadlocks. Even though it is called a fuzzer, it more resembles a record-analyze-replay mechanism.

Magic fuzzer is based on [DeadlockFuzzer](./DeadlockFuzzer).

### Magic Lock
To detect a potential cyclic deadlock in a recorded run, a lock order graph is build.

Before this graph is used to check for cycles, the
graph is first simplified.
First, magic fuzzer recognizes, that each node in the lock order graph that is part of a cycle must have at least one incoming and one outgoing edge. It therefore removes all nodes, that do not fulfill this
requirements. Removing those nodes may result in new nodes not having at least one incoming and one outgoing node. This pruning is therefore
iteratively applied until no more nodes can be removed.

This simplification seems to be the main advancement over [DeadlockFuzzer](./DeadlockFuzzer).

Then MagicLocker partitions the set of lock dependencies
by threads, sorts the partitions in the ascending order of their
thread identifiers to align its search sequence among the
partitions with the permutation of every potential cycle that a
thread with a smaller identifier always appears first in the
permutation. Because each thread can only occur once in a
cycle, Magiclock further employs a depth-first-search to
avoid exploring any subtree if any node in the path from the
root node to the current node in the search tree has the thread
identifier of the root node of the subtree.

### Deadlock Confirmation
To confirm if a detected cycle is an actual deadlock, MagicFuzzer
reruns the program in a controlled manner.

MagicFuzzer uses an active random scheduler to check
against a set of cycles (denoted as $CycleSet$) reported by
Magiclock with each execution. $CycleSet$ is a set of
cycles reported by Magiclock. $ToBePaused$ is a set of
threads with each thread existing in some cycles in
CycleSet. $ToAcquire(t)$ represents a lock that $t$ wants to
acquire in its next statement. $Paused$ is a set of pairs of a
thread $t$ and a Cycle, which denotes that, when executing $p$, a
thread will be paused and added into $Paused$ if $\langle t, ToAcquire(t), Lockset(t)\rangle# belongs to a Cycle. $Enable$ is a set of threads that each has not terminated yet. We use
$stmt$ to denote a high-level instruction, such as an acquire or
release operation. We denote a call to execute a statement
stmt by $execute(stmt)$. The functions $pause(t)$ and
$resume(t)$ represent the actions to pause and resume $t$,
respectively.

Given a program $p$ and a $CycleSet$, MagicScheduler
firstly identifies $ToBePaused$ set by extracting all identical
threads abstractions from each Cycle in $CycleSet$.
 It then initializes $Paused$ to be empty,
$Lockset(t)$ to be empty for each thread $t$, and
$Enable$ to contain all threads in $p$.
When executing $p$, if $t$ is not in $ToBePaused$,
MagicScheduler allows $t$ to execute statements. Otherwise, if
the next statement of $t$ is a lock acquisition statement
($acquire(t, m)$), just before executing this statement,
MagicScheduler checks whether any real deadlock may
occur if t acquires $m$ by calling $CheckDeadlock$. This function checks whether a real
deadlock occurs, and reports a deadlock if there exists a
cyclic lock dependency chain. No
matter if a deadlock occurs or not, MagicScheduler then calls
$CheckAndPause$ to determine whether or not the current
thread should be paused. If $CheckAndPause$ returns a Cycle,
MagicScheduler pauses $t$ and adds the pair $\langle t, Cycle\rangle$ into $Paused$; otherwise, MagicScheduler calls $execute(stmt)$ to
execute the statement, and updates the lockset of $t$. If the
statement $stmt$ is a lock release statement, MagicScheduler
updates the lockset of $t$ and calls $execute(stmt)$. All other
statements will be directly executed.