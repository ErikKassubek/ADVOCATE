# Order Based Fuzzing

The order based fuzzing is based on [GoPie](../relatedWorks/goPie.md).

## Idea
The main idea for [GoPie](https://github.com/CGCL-codes/GoPie) is to extend [GFuzz](https://github.com/system-pclub/GFuzz). Instead of just
influencing the select, it tries to influence the interleaving of all concurrent
operations in order to detect new bugs. GoPie uses only fork, channel and mutex operations.

## GoPie

### Feedback Collection
Not all run executions are again selected for mutations. The selection mainly
depends on whether the run was successful. Meaning, if the mutation replay
was interrupted by a timeout on one of the waiting operations being triggered
or if the runtime of a mutation exceeds a predefined values (e.g. 7 min per test).

### Fragmentation
To prevent the path explosion problem, GoPie divides the program into
smaller fragments, always only mutating one of them.

Those fragments are called scheduling chain (SC). In the original paper, they are defined as $$SC: \{\langle r_1, p_1, o_1\rangle\, ..., \langle r_i, p_i, o_i\rangle\, ..., \langle r_n, p_n, o_n\rangle\ |\ r_i \neq r_{i+1}, 1 \leq i \leq n-1\},$$
where each triplets refers to one operations. $r_i$ is the routine, where the operation took place.
$p_i$ is the type of primitive, e.g. channel or mutex. $o_n$ is an identifier for the operation, e.g. the line number. In our implementation, we build the chains directly as a slice of trace elements.

The chain stands for the order, in which the operations have been executed.
Additionally, two consecutive triples in a chain must belong to different routines. This means,
we only mutation the runs at positions, where the execution changes the routine. To build them, we traverse the trace in order of tPost and create sets of maximal length of consecutive operations, where two neighboring operations (we only look at the operations that are used for GoPie) are not in the same routine.

### Relations
We say $\langle c, c'\rangle \in CPOP_1$ is $c$ and $c_1$ are neighboring operations in the same routine. We say $\langle c, c' \rangle \in CPOP_2$ if $c$ and $c'$ are operations on the same primitive but in different routines.

To see which mutations on a chain are possible, GoPie defines two Relation between operations. Those relations are defined by the following rules:

- $Rule1$: $\exists c, c' \langle c, c' \rangle \in CPOP_1 \to c' \in Rel_1(c)$
- $Rule2$: $\exists c, c' \langle c, c' \rangle \in CPOP_2 \to c' \in Rel_2(c)$
- $Rule3$: $\exists c, c', c'', c' \in Rel_1(c), c'' \in Rel_2(c')\to c'' \in Rel_2(c)$
- $Rule4$: $\exists c, c', c'', c' \in Rel_2(c), c'' \in Rel_2(c')\to c'' \in Rel_2(c)$

where $Rel_{1/2}(x)$ stands for the set in which the operations of primitives are related to $x$.\
Rule 1 indicates, that two operations executed consecutively in the same routine are related.\
Rule 2 indicates, that two operations of the same primitive that execute consecutively in different routines are related.\
Rule 3 and 4 perform transitive inference.

Those relations are later used in the mutator in rule 3 and 4 (see below). They therefore
need to be calculated (do all of them need to be calculated up front or can
we only calculate those when needed $\to$ The transitive inference seems to make it necessary,
that all $Rel_{1/2}$ are calculated up front).

### Mutation
Given such a scheduling chain, it can be mutated with the following rules:

1. Abridge: This removes an item from the $SC$ (either from head or tail) if there
is more than one operation in it, which helps GoPie to limit the length of an $SC$. $$\exists o_i, o_j \{o_i, o_{i+1},...,o_{j-1}, o_j\} \in SC \to \{o_{i+1},...,o_{j-1}, o_j\} \in SC, \{o_i, o_{i+1},...,o_{j-1}\} \in SC$$
2. Flip: This performes a reverse process on the $SC$ to be mutated. E.g. if $\langle s_1, s_2\rangle$ is covered in the scheduling, $\langle s_2, s_1\rangle$ is also valuable to take a try $$\exists o_i, o_j \{...,o_i, o_j,...\}\in SC \to \{...,o_j, o_i,...\}\in SC$$
3. Substitute: This tries to replace an operation with another one from the set of $Rel_1$ $$\exists o_i, o_j \in Rel_1(o_i), \{...,o_i,...\} \in SC \to \{...,o_j,...\} \in SC$$
4. Augment: This tries to increase the length of $SC$ by adding another operation from the set of $Rel_2$ to its tail, which aims to explore those effective interleaving in a further step. $$\exists o_j, o_j \in Rel_2(o_i), \{...,o_j\}\in SC \to \{...,o_j, o_j\}\in SC$$

Given a chain `c`, we construct the new mutated chains as follows

```go
res := make(map[string]chain)
res[c.toString()] = c

for {
  for _, ch := range res {
    tset := make(map[string]chain, 0)

    // Rule 1 -> abridge
    if ch.len() >= 2 {
      newCh1, newCh2 := abridge(ch)
      tset[newCh1.toString()] = newCh1
      tset[newCh2.toString()] = newCh2
    }

    // Rule 2 -> flip (not in original implementation)
    if ch.len() >= 2 {
      newChs := flip(ch)
      for _, newCh := range newChs {
        tset[newCh.toString()] = newCh
      }
    }

    // Rule 3 -> substitute
    if ch.len() <= BOUND && rand.Int()%2 == 1 {
      newChs := substitute(ch)
      for _, newCh := range newChs {
        tset[newCh.toString()] = newCh
      }
    }

    // Rule 4 -> augment
    if ch.len() <= BOUND && rand.Int()%2 == 1 {
      newChs := augment(c)
      for _, newCh := range newChs {
        tset[newCh.toString()] = newCh
      }
    }

    for k, v := range tset {
      res[k] = v
    }

    if len(res) > MUTATEBOUND {
      break
    }

    if (rand.Int() % 200) < energy {
      break
    }
  }
}
```
with `BOUND = 3` and `MUTATEBOUND = 138`. The implementation of the energy value seems not to be fully completed in the original implementation. The original value is calculated by a score function based on the interleaving. <!--TODO: describe more--> But since this value is basically always greater than 100, but then when starting a mutation capt to 100, the value seems to always be 100. I assume the main idea, that was never implemented for the goPie was to reduce this energy value with the number of already created mutation, to limit the total number of mutations. For our implementation the energy value is, for now, always just set to 100. In the future, this may change.

For now, we create a new replay for each mutated scheduling chain. For the future it may make sense, to merge multiple mutated scheduling chains into one execution, to reduce the required number of runs.

<!-- ### GoPie implementation of mutation
Info: mutator in file GoPie/pkg/fuzzer/mutator.go:42
- Rule 2 doesn't seem to be implemented
- The creation of new mutations seems to stop, if after iterating over all
already created mutations the number of total mutations is greater then 128
or if not then with a probability of 50% (Why?)
  - There is an energy value that should somehow control it, but it is just set
  to a constant and has TODOs everywhere
- Not clear how $SC$ are chosen (maybe all are mutated at the same time?)
 -->


### Order Enforcement
For the order enforcement, we use the order enforcement implemented for the [replay](../replay.md) mechanism. This means we write a trace file and replay it. For now, the replay files are created as follows: For each modified scheduling chain, we create one replay trace. For each of them, we reorder the elements in the recorded trace that correspond to the elements in the scheduling chain, to follow the order from the chain. We then remove all elements after the last element in the modified scheduling chain and add a replay end trace element. This allows us to force the program to execute the scheduling chain and then run without any guidance, hopefully executing new code.

## Improvements

> [!WARNING]
> This is not fully implemented yet

Based on the original paper, we now implement multiple improvements.

### HB

### Mutation

A downside of this approach is, that it creates a lot of mutations, that are
not possible to be executed.

Let' e.g. look at the following simple example:
```go
m := sync.Mutex{}

go func() {
  m.Lock()     // l1
  m.Unlock()   // u1
}

m.Lock()       // l2
m.Unlock()     // u2
```

We assume the execution order l1,u1,l2,u2.
Since the execution changes the routine between u1 and l2, they are
part of one execution chain. A mutation could therefore flip those two events,
creating l1,l2,u1 as a schedule. This schedule is not valid, since the
l2 would try to lock a mutex that is already locked. This would lead to
the program getting stuck or, with timeouts, to have the same execution
as the originally recorded version.

The original goPie implementation has no way to detect those impossible schedules
and will replay them, resulting in an stuck replay. GoPie will not create
new mutations from those failed runs or will be able to draw any conclusion from
them, but since those schedules are still replayed, they increase the number
of runs and therefore the runtime.

We can use our HB relation to directly check for schedules, that
violate the relation, To do this, we first create all mutation from a given schedule.
We then traverse all proposed schedules and remove all impossible once.

This noticeably reduces the number of created mutations and therefore the runtime.

### Analysis

GoPie can only detect concurrency bugs, when the bugs are directly triggered.
Using our predictive analysis, we can detect bugs even if they are not
directly triggered.

### Pre
TODO: write

