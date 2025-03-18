# Order Based Fuzzing

The order based fuzzing is based on [GFuzz](https://github.com/system-pclub/GFuzz).

> [!WARNING]
> This is not implemented yet

## Idea
The main idea for [GoPie](https://github.com/CGCL-codes/GoPie) is to extend [GFuzz](https://github.com/system-pclub/GFuzz). Instead of just
influencing the select, it tries to influence the interleaving of all concurrent
operations in order to detect new bugs.

## Feedback Collection
Not all run executions are again selected for mutations. The selection mainly
depends on whether the run was successful. Meaning, if the mutation replay
was interrupted by a timeout on one of the waiting operations being triggered
or if the runtime of a mutation exceeds a predefined values (e.g. 7 min per test).

## Fragmentation
To prevent the path explosion problem, GoPie divides the program into
smaller fragments, always only mutating one of them. For this,
it is determined, wether two primitives are related and
unrelated primitives are then separated into separate sets. A localized
analysis/mutation is then conducted on each of those sets.

GoPie defines the following rules, to determine wether two primitives are related

- $Rule1$: $\exists c, c' \langle c, c' \rangle \in CPOP_1 \to c' \in Rel_1(c)$
- $Rule2$: $\exists c, c' \langle c, c' \rangle \in CPOP_2 \to c' \in Rel_2(c)$
- $Rule3$: $\exists c, c', c'', c' \in Rel_1(c), c'' \in Rel_2(c')\to c'' \in Rel_2(c)$
- $Rule4$: $\exists c, c', c'', c' \in Rel_2(c), c'' \in Rel_2(c')\to c'' \in Rel_2(c)$

where $Rel_{1/2}(x)$ stands for the set in which the operations of primitives are related to $x$.\
Rule 1 indicates, that two operations executed consecutively in the same routine are related.\
Rule 2 indicates, that two operations of the same primitive that execute consecutively in different routines are related.\
Rule 3 and 4 perform transitive inference.

Those rel are later used in the mutator in rule 3 (see below). They therefore
need to be calculated (do all of them need to be calculated up front or can
we only calculate those when needed $\to$ The transitive inference seems to make it necessary,
that all $Rel_{1/2}$ are calculated up front).

As an example, the GoPie paper has the following code

```go
 1 type statusManager struct {
 2   podStatusesLock sync.Mutex
 3   podStatusChannel chan bool
 4   // ...
 5 }
 6
 7 func(s *statusManager)Start() {
 8   for i:=0; i<2;i++{
 9     <-s.podStatusChannel
10     s.podStatusesLock.Lock()
11     // handle the pod status here
12     s.podStatusesLock.Unlock()
13   }
14 }
15
16 func(s *statusManager)SetPodStatus() {
17   s.podStatusesLock.Lock()
18   // send the pod status below
19   s.podStatusChannel <- true
20   s.podStatusesLock.Unlock()
21 }
22
23 func main(){
24   s := &statusManager{podStatusChannel: make(chan bool)} // unbuffered
25   go s.Start()         // G1
26   go s.SetPodStatus()  // G2
27   go s.SetPodStatus()  // G3
28 }
```

With the following run\
<img src="./img/execution.png" width="600">

According to the paper, this results in the following relations\
<img src="./img/relations.png" width="400">


## Interleaving Mutators
GoPie defines scheduling chains (SC) as $$SC: \{\langle r_1, p_1, o_1\rangle\, ..., \langle r_i, p_i, o_i\rangle\, ..., \langle r_n, p_n, o_n\rangle\ |\ r_i \neq r_{i+1}, 1 \leq i \leq n-1\},$$
where each triplets refers to one operations. $r_i$ is the routine, where the operation took place.
$p_i$ is the type of primitive, e.g. channel or mutex. $o_n$ is an identifier for the operation, e.g. the line number. The chain stands for the order, in which the operations have been executed.
Additionally, two consecutive triples in a chain must belong to different routines. This means,
we only mutation the runs at positions, where the execution changes the routine.

Given such a scheduling chain, it can be mutated with the following rules:

1. Abridge: This removes an item from the $SC$ (either from head or tail) if there
is more than one operation in it, which helps GoPie to limit the length of an $SC$. $$\exists o_i, o_j \{o_i, o_{i+1},...,o_{j-1}, o_j\} \in SC \to \{o_{i+1},...,o_{j-1}, o_j\} \in SC, \{o_i, o_{i+1},...,o_{j-1}\} \in SC$$
2. Flip: This performes a reverse process on the $SC$ to be mutated. E.g. if $\langle s_1, s_2\rangle$ is covered in the scheduling, $\langle s_2, s_1\rangle$ is also valuable to take a try $$\exists o_i, o_j \{...,o_i, o_j,...\}\in SC \to \{...,o_j, o_i,...\}\in SC$$
3. Substitute: This tries to replace an operation with another one from the set of $Rel_1$ $$\exists o_i, o_j \in Rel_1(o_i), \{...,o_i,...\} \in SC \to \{...,o_j,...\} \in SC$$
4. Augment: This tries to increase the length of $SC$ by adding another operation from the set of $Rel_2$ to its tail, which aims to explore those effective interleaving in a further step.

In the implementation of GoPie, the order of testing the potential interleavings is randomized.

### GoPie implementation of mutation
Info: mutator in file GoPie/pkg/fuzzer/mutator.go:42
- Rule 2 doesn't seem to be implemented
- The creation of new mutations seems to stop, if after iterating over all
already created mutations the number of total mutations is greater then 128
or if not then with a probability of 50% (Why?)
  - There is an energy value that should somehow control it, but it is just set
  to a constant and has TODOs everywhere
- Not clear how $SC$ are chosen (maybe all are mutated at the same time?)



## Order Enforcement
The order enforcement for the code is similar to the old way the replay
order enforcement was implemented. Before it executes and execution if
first checks whether the execution is the next execution in the queue to be
executed. If this is not the case, it will wait with an loop, periodically
checking again, until either the operation is the next operation to be executed
or until a timeout has run out.

In general, the replay implemented in ADVOCATE should be able to handle the
order enforcement.