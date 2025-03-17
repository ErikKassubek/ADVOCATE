# Select

Channel operations including select statements are recorded in the trace where the select statement is located.

## Info
Select statements with only one case are not recorded as select statements. If the case is an default case, it is not recorded at all. If the case is a send or receive, the select statement is equivalent to just the send/receive statement and therefor recorded as such.


## Trace element
The basic form of the trace element is
```
S,[tpre],[tpost],[id],[cases],[selIndex],[pos]
```
where `S` identifies the element as a select element.
The other fields are set as follows:
- [tpre] $\in \mathbb N$: This is the value of the global counter when the routine gets to the select statement.
- [tpost] $\in \mathbb N$: This is the value of the global counter when the select has finished, meaning if either on of the select cases has successfully send or received or the default case has been executed
- [id]: This field contains the unique id of the select statement
- [cases]: This field shows a list of the available cases in the select. The
cases are separated by a ~ symbol. The elements are equal to an equivalent
channel operation without the `pos` field, except that the fields are separated
by a decimapl point (.) instead of a comma. . The operations are ordered in the following way: First all send
operations, ordered as in the order of the select cases, then all receive operations,
again ordered as written in the select cases. If a case is on a nil channel, the channel ID is set to *.
By checking the the `tpre` of those channel operations are
all the same as the `tpre` of the select. The `tpost` can be used to find out,
which case was selected. For the selected and executed case, `tpost` is equal
to the `tpost` of the select statement. For all other cases it is zero. The channel
element also includes the operation id `oId`, to connect the sending and
receiving operations. If the
select contains a default case, it is denoted with a single `d` as the last
element in the cases list. If the default value was chosen it is capitalized (`D`).
- [selIndex]: The internal index of the selected case.
- [pos]: The last field show the position in the code, where the select is implemented. It consists of the file and line number separated by a colon (:). It the select only contains one case, the line number is
equal to the line of this case.

A select statement is only recorded as such, if it contains at least two non-default cases. Otherwise the Go compiler rewrites it
internally as a normal channel operation and is therefore recorded as such. A select case with one non-default and a default case is
only recorded, if the non-default case was chosen.

## Implementation
The implementation for the select has been split in the implementation for the [original select](../../go-patch/src/runtime/select.go#L629) and the [select with preferred case](../../go-patch/src/runtime/select.go#L151) (for more info see [here](../replay.md#select)). The entry point for select in the [selectgo](../../go-patch/src/runtime/select.go#L123) function now calls one of the two select versions. The recording is done in the two selects separately, since we need to no the lock order of the cases. TODO: continue