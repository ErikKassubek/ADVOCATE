# Effective Concurrency Testing for Go via Directional Primitive-Constrained Interleaving Exploration

[Z. Jiang, M. Wen, Y. Yang, C. Peng, P. Yang and H. Jin, "Effective Concurrency Testing for Go via Directional Primitive-Constrained Interleaving Exploration," 2023 38th IEEE/ACM International Conference on Automated Software Engineering (ASE), Luxembourg, Luxembourg, 2023, pp. 1364-1376, doi: 10.1109/ASE56229.2023.00086.](https://dl.acm.org/doi/10.1109/ASE56229.2023.00086)

## Summary
The main idea of this paper is to extend [GFuzz](./GFuzz.md). Instead of just
influencing the select, it tries to influence the interleaving of all concurrent
operations in order to detect new bugs. The implemented tool GoPie modifies fork, channel and mutex operations.


## Fragmentation
To prevent the path explosion problem, GoPie divides the program into
smaller fragments, always only mutating one of them.

Those fragments are called scheduling chain (SC). In the original paper, they are defined as $$SC: \{\langle r_1, p_1, o_1\rangle\, ..., \langle r_i, p_i, o_i\rangle\, ..., \langle r_n, p_n, o_n\rangle\ |\ r_i \neq r_{i+1}, 1 \leq i \leq n-1\},$$
where each triplets refers to one operations. $r_i$ is the routine, where the operation took place.
$p_i$ is the type of primitive, e.g. channel or mutex. $o_n$ is an identifier for the operation, e.g. the line number. In our implementation, we build the chains directly as a slice of trace elements.

The chain stands for the order, in which the operations have been executed.
Additionally, two consecutive triples in a chain must belong to different routines. This means,
we only mutation the runs at positions, where the execution changes the routine. To build them, we traverse the trace in order of tPost and create sets of maximal length of consecutive operations, where two neighboring operations (we only look at the operations that are used for GoPie) are not in the same routine.

## Relations
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

## Mutation
Given such a scheduling chain, it can be mutated with the following rules:

1. Abridge: This removes an item from the $SC$ (either from head or tail) if there
is more than one operation in it, which helps GoPie to limit the length of an $SC$. $$\exists o_i, o_j \{o_i, o_{i+1},...,o_{j-1}, o_j\} \in SC \to \{o_{i+1},...,o_{j-1}, o_j\} \in SC, \{o_i, o_{i+1},...,o_{j-1}\} \in SC$$
2. Flip: This performes a reverse process on the $SC$ to be mutated. E.g. if $\langle s_1, s_2\rangle$ is covered in the scheduling, $\langle s_2, s_1\rangle$ is also valuable to take a try $$\exists o_i, o_j \{...,o_i, o_j,...\}\in SC \to \{...,o_j, o_i,...\}\in SC$$
3. Substitute: This tries to replace an operation with another one from the set of $Rel_1$ $$\exists o_i, o_j \in Rel_1(o_i), \{...,o_i,...\} \in SC \to \{...,o_j,...\} \in SC$$
4. Augment: This tries to increase the length of $SC$ by adding another operation from the set of $Rel_2$ to its tail, which aims to explore those effective interleaving in a further step. $$\exists o_j, o_j \in Rel_2(o_i), \{...,o_j\}\in SC \to \{...,o_j, o_j\}\in SC$$

## Feedback Collection
Not all run executions are again selected for mutations. The selection mainly depends on whether the run was successful. Meaning, if the mutation replay was interrupted by a timeout on one of the waiting operations being triggered or if the runtime of a mutation exceeds a predefined values (e.g. 7 min per test).

## Order Enforcement
The order enforcement is similar to the enforcement used for ADVOCATE.
An operation will only execute if it is the next element in the schedule. Otherwise it will wait.

<center><img src="../../../img/relatedWorkGoPieOrder.png" alt="Order enforcement" width="400px" height=auto></center>