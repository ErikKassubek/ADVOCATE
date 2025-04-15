# Greybox Fuzzing for Concurrency Testing

[D. Wolff, Z. Shi, G. J. Duck, U. Mathur, and A. Roychoudhury, “Greybox fuzzing for con-
currency testing,” in Proceedings of the 29th ACM International Conference on Architec-
tural Support for Programming Languages and Operating Systems, Volume 2, ser. ASPLOS
’24, La Jolla, CA, USA: Association for Computing Machinery, 2024, pp. 482–498, isbn:
9798400703850. doi: 10.1145/3620665.3640389. [Online]. Available: https://doi.org/10.1145/3620665.3640389.](https://dl.acm.org/doi/10.1145/3620665.3640389)

## Summary
The paper employs a biased random search which guides exploration towards neighborhoods which will likely expose new behavior. To
identify new behaviors, it uses the "reads-from" relation.

The Paper uses a grey box fuzzing approach.

<center><img src="../img/relatedWorksGreyboxFuzzingAlgo.png" alt="Greybox Fuzzing" width="400px" height=auto></center>


### Read-from and (abstract) schedules

The read-from function $rf_\sigma$ maps each read event $e$ to its corresponding write event $f = rf_\sigma(e)$ from which it observes its value from.\
Two schedules $\sigma_1$ and $\sigma_2$ are read-from equivalent ($\sigma_1 \equiv_{rf} \sigma_2$), if they both observe the same events and the same read-from function. The paper assumes, that if $\sigma_1 \equiv_{rf} \sigma_2$, $\sigma_2$ must observe the same control flow and thus identically expose any given bug.

To reduce the search space of program interleaving, schedules are represented by there $\equiv_{rf}$-equivalence class rather than a concrete sequence of events. Those abstract schedules are then mutated. An abstract schedule $\alpha = \alpha^+ \cup \alpha^-$ is a set of positive and negative reads-from constraints $\{C_1^+,...,c_{n_1}^+, C_1^-,...,c_{n_1}^-\}$, where $C_i^+ = ea_i \overset{rf}\to ea'_i$ and $C_i^+ = ea_i \overset{rf}{\not\to} ea'_i$ for some read $aa_i$ and write $ea'_i$ abstract events.

A (concrete) schedule $\sigma$ is an instantiation of an abstract schedule $\alpha$ if it satisfies the constraints of $\alpha$, i.e.,
1. for every positive constraint $C^+_𝑖 = ea_i \overset{rf}{\to} ea'_i \in\alpha$, there are events $e_i$ and $e'_i$ in $\sigma$ that instatiate $ea_i$ and $ea'_i$ respectively and satisfy $rf_\sigma(e') = e_i$ , and
2. for every negative constraint $C^-_𝑖 = ea_i \overset{rf}{\not \to} ea'_i \in\alpha$, there are no events $e_i$ and $e'_i$ in $\sigma$ that instatiate $ea_i$ and $ea'_i$ respectively and satisfy $rf_\sigma(e') = e_i$.

If $\sigma_1 \equiv_{rf} \sigma_2$, then either both or none of them are an instantiation fo any given abstract schedule $\alpha$


## Mutation

The paper now mutates these abstract schedules. First it randomly chooses one of the four mutation operators:

- $INSERT(\alpha, C) = \alpha \cup \{C\}$
- $SWAP(\alpha, C_1, C_2) = (\alpha \setminus \{c_1\} \cup \{C_2\})$
- $DELETE(\alpha, C) = \alpha \setminus \{C\}$
- $NEGATE(\alpha, C) = SWAP(\alpha, C, \lnot C)$

It them randomly picks potentially conflicting events from the set of all events observed, to form constraints $C_1,...,C_n$ which are needed by these mutation operators.

## Is Interesting
A mutation is only executed if it is deemed interesting. A mutation is interesting, if there is a reads-from pair $(e_1, e_2)$ in $\sigma_{mut}$, that no other of the abstract schedules in $S$ instantiates it, or if the schedule results in a crash. Additionally, a cut-off exponential power-schedule is used, to further bias exploration towards rarely used constraints.

## Scheduling

Given such an abstract schedule, it is not trivial (and may be impossible) to find a concrete execution which satisfies
those scheduling constraints.

The scheduling algorithm works by determining which events
to execute next when multiple enabled events are possible.
To push the execution towards satisfying abstract scheduling constraints, a greedy scheduling algorithm is used to
delay or immediately execute events involved in these constraints. Such a priority change ensures that the event will
always (or never) be chosen next over other enabled events
that are not in the abstract schedule.

For example, to satisfy $w \overset{rf}{\to}$, we can boost the priority of a write event $e_w$ that instantiates the abstract event $w$, and then subsequently a read event $e_r$ that instantiates $r$, to satisfy the reads-from constraint.

However, simply boosting and lowering the priorities when the relevant events are both enabled is often
inadequate to ensure that a desired abstract schedule constraint is met. Relevant events are often widely separated
in execution traces and thus likely will not be simultane-
ously enabled without further intervention. Therefore, a
succinct state machine for each constraint is maintained in the
abstract schedule to determine how priorities of relevant
events should be proactively adjusted.

<center><img src="../img/relatedWorksGreyboxFuzzingState.png" alt="State machines" width="900px" height=auto></center>

In the absence of relevant events, the paper first assigns each event a random score if it does not already have one. It then picks the event with the highest score to execute next, resetting that event’s score along with the scores of any racing events.