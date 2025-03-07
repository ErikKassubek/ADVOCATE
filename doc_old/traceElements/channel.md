# Channel
The sending and receiving on and the closing of channels is recorded in the
trace of the routine where they occur.

## Trace element
The basic form of the trace element is
```
C,[tpre],[tpost],[id],[opC],[cl],[oId],[qSize],[qCount],[pos]
```
where `C` identifies the element as a channel element. The other fields are
set as follows:
- [tpre] $\in \mathbb N$ : This is the value of the global counter when the operation starts
the execution of the operation
- [tpost]$\in \mathbb N$: This is the value of the global counter when the channel has finished its operation. For close we get [tpost] = [tpre]
- [id]$\in \mathbb N$: This shows the unique id of the channel. If the channel is nil, this is `*`
- [opC]: This field shows the operation that was executed:
    - [opC] = `S`: send
    - [opC] = `R`: receive
    - [opC] = `C`: close
- [cl]: If this value is set to `t`, the operation was finished, because the channel was closed in another routine, while or before the channel was waiting at this operation. This means, that the operation was never really executed, even though tpost is not zero.
This can only exist for send or receive. In other cases, this is `f`.
- [oId] $\in \mathbb N$: This field shows the communication id. This can be used to connect corresponding communications. If a send and a receive on the same channel (same channel id) have the same [oId], a message was send from the send to the receive. For close this is always `0`
- [qSize] $\in \mathbb N_0$: This is the size of the channel. For unbuffered channels this is `0`.
- [qCount] $\in \mathbb N_0$: Number of elements in the queue after the operations was executed
- [pos]: The last field show the position in the code, where the mutex operation
was executed. It consists of the file and line number separated by a colon (:)
## Example
The following is an example for a program with different channel operations:
```go
package main
func main() {    // routine 1
    c := make(chan int, 2) // id = 4
	d := make(chan int, 0) // id = 5

	go func() { // routine 2
		c <- 1 // line 7
		c <- 2 // line 8
		d <- 1 // line 9
		d <- 1 // line 10
	}()

	<-d // line 12
	<-c // line 13
	<-c // line 14

	close(c) // line 16
}
```
If we ignore all internal operations, we would get the following trace:
```txt
G,1,2;C,2,10,5,R,f,1,0,example_file.go:12;C,11,12,4,R,f,1,2,example_file.go:13;C,13,14,4,R,f,2,2,example_file.go:14;C,15,15,4,C,f,0,2,example_file.go:16
```
```txt
C,3,4,4,S,f,1,2,example_file.go:7;C,5,6,4,S,f,2,2,example_file.go:8;C,7,8,5,S,f,1,0,example_file.go:9;C,9,0,5,S,f,2,0,example_file.go:10
```
In this example it is also shown what happens, when an operation is not fully executed. In this case [tpre] is set, but [tpost] is the default value of 0 (last element in trace, line 10).

## Implementation
The recording of the channel operations is done in the
`go-patch/src/runtime/chan.go` file in the `chansend`, `chanrecv` and `closechan` function. Additionally the
`hchan` struct in the same file is ammended by the following fields:
- `id`: identifier for the channel
- `numberSend`: number of completed send operations
- `numberRecv`: number of completed reveive operations

`numberSend` and `numberRecv` are later set as `oId` in the corresponding trace elements. The send operations are implemented as a FIFO-queue. We can therefore count the number of elements added to the queue and removed from the
queue, to determine, which send and receive operations are
communication partners. Because of mutexes, that are already present in the original channel implementation,
it is not possible to mix up these numbers.\
For the send and receive operations three record functions are added. The first one (`DedegoChanSendPre`/`DedegoChanRecvPre`) at the beginning of the operation, which records [tpre], [id], [opC], [qSize] and [pos].\
The other two functions are called at the end of the
operation, after the send or receive was fully executed.
These functions record [tpost] (`DedegoChanPost`).\
As a close on a channel cannot block, it only needs one recording function. This function (`DedegoChanClose`) records all needed values. For [tpre] and [tpost] the same
value is set. The same is true for [qCountPre] and [qCountPost].