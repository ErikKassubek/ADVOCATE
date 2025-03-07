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

## Example
The following is an example containing two select statements, one with and one without a select:
```go
package main
func main() {  // Routine 1
    c := make(chan int, 0)  // id = 1
	d := make(chan int, 0)  // id = 2
	e := make(chan int, 1)  // id = 3

	go func() {  // routine 2
		select {  // id = 4, line 8
		case <-d:
			println("d1")
		case <-e:
			println("e1")
		default:
			println("default")
		}

		c <- 1  // line 17
	}()

	select {  // id = 5, line 20
	case <-c:
		println("c2")
	case <-d:
		println("d2")
	}
}
```
If we ignore all internal operations, we get the following trace:
```txt
G,1,2;S,3,8,5,C.3.0.2.R.f.0.0~C.3.8.1.R.f.1.0,1,example_file.go:20
```
```txt
S,4,5,4,C.4.0.3.R.f.0.1~C.4.0.2.R.f.0.0~D,-1,example_file.go:8;C,6,7,1,S,1,0,example_file.go:17
```

## Implementation
The recording of the select statement is done in the `selectgo` function in the `go-patch/src/runtime/select.go` file. It contains two function calls. The first one is called shortly after the beginning of the `selectgo` functions to record the available cases and there internal order. The second function call is done when `selectgo` returns, to record
the successful execution of the statement and the selected case.\
Selects with only one case are automatically turned into a pure
send or receive by the compiler. For this reason, these cases must
not be additionally recorded.\
Selects with exactly one non-default and one default case are als
rewritten by the compiler. In this case it is necessary to record
these cases, because the rewritten version does not use the `select`
go function. For this reason, two additional recorder functions
have been added to the `selectnbsend` and `selectnbrecv` functions
in the `go-patch/src/runtime/chan.go` file.