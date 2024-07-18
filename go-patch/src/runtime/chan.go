// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// This file contains the implementation of Go channels.

// Invariants:
//  At least one of c.sendq and c.recvq is empty,
//  except for the case of an unbuffered channel with a single goroutine
//  blocked on it for both sending and receiving using a select statement,
//  in which case the length of c.sendq and c.recvq is limited only by the
//  size of the select statement.
//
// For buffered channels, also:
//  c.qcount > 0 implies that c.recvq is empty.
//  c.qcount < c.dataqsiz implies that c.sendq is empty.

import (
	"internal/abi"
	"runtime/internal/atomic"
	"runtime/internal/math"
	"unsafe"
)

const (
	maxAlign  = 8
	hchanSize = unsafe.Sizeof(hchan{}) + uintptr(-int(unsafe.Sizeof(hchan{}))&(maxAlign-1))
	debugChan = false
)

type hchan struct {
	qcount   uint           // total data in the queue
	dataqsiz uint           // size of the circular queue
	buf      unsafe.Pointer // points to an array of dataqsiz elements
	elemsize uint16
	closed   uint32
	elemtype *_type // element type
	sendx    uint   // send index
	recvx    uint   // receive index
	recvq    waitq  // list of recv waiters
	sendq    waitq  // list of send waiters

	// lock protects all fields in hchan, as well as several
	// fields in sudogs blocked on this channel.
	//
	// Do not change another G's status while holding this lock
	// (in particular, do not ready a G), as this can deadlock
	// with stack shrinking.
	lock mutex

	// ADVOCATE-CHANGE-START
	id              uint64 // id of the channel
	numberSend      uint64 // number of completed send operations
	numberSendMutex mutex  // mutex for numberSend
	numberRecv      uint64 // number of completed recv operations
	numberRecvMutex mutex  // mutex for numberRecv
	advocateIgnore  bool   // if true, the channel is ignored by tracing and replay
	// ADVOCATE-CHANGE-END
}

type waitq struct {
	first *sudog
	last  *sudog
}

//go:linkname reflect_makechan reflect.makechan
func reflect_makechan(t *chantype, size int) *hchan {
	return makechan(t, size)
}

func makechan64(t *chantype, size int64) *hchan {
	if int64(int(size)) != size {
		panic(plainError("makechan: size out of range"))
	}

	return makechan(t, int(size))
}

func makechan(t *chantype, size int) *hchan {
	elem := t.Elem

	// ADVOCATE-CHANGE-START
	advocateIgnored := false
	if size == 1<<16 {
		advocateIgnored = true
		size = 0
	}
	// ADVOCATE-CHANGE-END

	// compiler checks this but be safe.
	if elem.Size_ >= 1<<16 {
		throw("makechan: invalid channel element type")
	}
	if hchanSize%maxAlign != 0 || elem.Align_ > maxAlign {
		throw("makechan: bad alignment")
	}

	mem, overflow := math.MulUintptr(elem.Size_, uintptr(size))
	if overflow || mem > maxAlloc-hchanSize || size < 0 {
		panic(plainError("makechan: size out of range"))
	}

	// Hchan does not contain pointers interesting for GC when elements stored in buf do not contain pointers.
	// buf points into the same allocation, elemtype is persistent.
	// SudoG's are referenced from their owning thread so they can't be collected.
	// TODO(dvyukov,rlh): Rethink when collector can move allocated objects.
	var c *hchan
	switch {
	case mem == 0:
		// Queue or element size is zero.
		c = (*hchan)(mallocgc(hchanSize, nil, true))
		// Race detector uses this location for synchronization.
		c.buf = c.raceaddr()
	case elem.PtrBytes == 0:
		// Elements do not contain pointers.
		// Allocate hchan and buf in one call.
		c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
		c.buf = add(unsafe.Pointer(c), hchanSize)
	default:
		// Elements contain pointers.
		c = new(hchan)
		c.buf = mallocgc(mem, elem, true)
	}

	c.elemsize = uint16(elem.Size_)
	c.elemtype = elem
	c.dataqsiz = uint(size)

	// ADVOCATE-CHANGE-START
	// get and save a new id for the channel
	c.advocateIgnore = advocateIgnored
	if !c.advocateIgnore {
		c.id = GetAdvocateObjectID()
	}
	// ADVOCATE-CHANGE-END

	lockInit(&c.lock, lockRankHchan)

	if debugChan {
		print("makechan: chan=", c, "; elemsize=", elem.Size_, "; dataqsiz=", size, "\n")
	}
	return c
}

// ADVOCATE-CHANGE-START
func (c *hchan) SetAdvocateIgnore() {
	c.advocateIgnore = true
}

// ADVOCATE-CHANGE-END

// chanbuf(c, i) is pointer to the i'th slot in the buffer.
func chanbuf(c *hchan, i uint) unsafe.Pointer {
	return add(c.buf, uintptr(i)*uintptr(c.elemsize))
}

// full reports whether a send on c would block (that is, the channel is full).
// It uses a single word-sized read of mutable state, so although
// the answer is instantaneously true, the correct answer may have changed
// by the time the calling function receives the return value.
func full(c *hchan) bool {
	// c.dataqsiz is immutable (never written after the channel is created)
	// so it is safe to read at any time during channel operation.
	if c.dataqsiz == 0 {
		// Assumes that a pointer read is relaxed-atomic.
		return c.recvq.first == nil
	}
	// Assumes that a uint read is relaxed-atomic.
	return c.qcount == c.dataqsiz
}

// entry point for c <- x from compiled code.
//
//go:nosplit
func chansend1(c *hchan, elem unsafe.Pointer) {
	// ADVOCATE-CHANGE-START
	chansend(c, elem, true, getcallerpc(), false)
	// ADVOCATE-CHANGE-END
}

/*
 * generic single channel send/recv
 * If block is not nil,
 * then the protocol will not
 * sleep but return if it could
 * not complete.
 *
 * sleep can wake up with g.param == nil
 * when a channel involved in the sleep has
 * been closed.  it is easiest to loop and re-run
 * the operation; we'll see that it's now closed.
 */
// ADVOCATE-CHANGE-START
// set ignored to true, if it is used in a one case + default select. In this case, it is recorded and replayed in the select
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr, ignored bool) bool {
	// ADVOCATE-CHANGE-END
	if c == nil {
		if !ignored {
			AdvocateChanSendPre(0, 0, 0, true)
		}
		if !block {
			return false
		}
		gopark(nil, nil, waitReasonChanSendNilChan, traceBlockForever, 2)
		throw("unreachable")
	}

	if debugChan {
		print("chansend: chan=", c, "\n")
	}

	if raceenabled {
		racereadpc(c.raceaddr(), callerpc, abi.FuncPCABIInternal(chansend))
	}

	// ADVOCATE-CHANGE-START
	// wait until the replay has reached the current point
	var replayElem ReplayElement
	var enabled bool
	var valid bool
	if !ignored && !c.advocateIgnore {
		enabled, valid, replayElem = WaitForReplay(OperationChannelSend, 3)
		if enabled && valid {
			if replayElem.Blocked {
				lock(&c.numberSendMutex)
				c.numberSend++
				unlock(&c.numberSendMutex)
				_ = AdvocateChanSendPre(c.id, c.numberSend, c.dataqsiz, false)
				BlockForever()
			}
		}
	}
	// ADVOCATE-CHANGE-END

	// Fast path: check for failed non-blocking operation without acquiring the lock.
	//
	// After observing that the channel is not closed, we observe that the channel is
	// not ready for sending. Each of these observations is a single word-sized read
	// (first c.closed and second full()).
	// Because a closed channel cannot transition from 'ready for sending' to
	// 'not ready for sending', even if the channel is closed between the two observations,
	// they imply a moment between the two when the channel was both not yet closed
	// and not ready for sending. We behave as if we observed the channel at that moment,
	// and report that the send cannot proceed.
	//
	// It is okay if the reads are reordered here: if we observe that the channel is not
	// ready for sending and then observe that it is not closed, that implies that the
	// channel wasn't closed during the first observation. However, nothing here
	// guarantees forward progress. We rely on the side effects of lock release in
	// chanrecv() and closechan() to update this thread's view of c.closed and full().
	if !block && c.closed == 0 && full(c) {
		return false
	}

	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
	}

	lock(&c.lock)

	// ADVOCATE-CHANGE-START
	// this block is called if a send is made on a channel
	// it increases the number of sends on the channel, which is used to
	// identify the communication partner in the advocate analysis
	// After that a channel send event is created in the trace to show,
	// that the channel tried to send.
	// The current function 'chansend' only returns, if the send was successful,
	// meaning the channel either directly communicated with a receive or wrote
	// into the channel buffer. Therefor, the send event is modified to include
	// the post information by AdvocateChanPost, if 'chansend' returns.
	// advocateIndex is used to connect the post event to the correct
	// pre envent in the trace.
	var advocateIndex int
	if !ignored && !c.advocateIgnore {
		lock(&c.numberSendMutex)
		c.numberSend++
		unlock(&c.numberSendMutex)
		advocateIndex = AdvocateChanSendPre(c.id, c.numberSend, c.dataqsiz, false)
	}
	// ADVOCATE-CHANGE-END

	if c.closed != 0 {
		unlock(&c.lock)
		if enabled {
			IsNextElementReplayEnd(ExitCodeSendClose, true, false)
		}

		// ADVOCATE-CHANGE-START
		if !ignored && !c.advocateIgnore {
			AdvocateChanPostCausedByClose(advocateIndex)
		}
		// ADVOCATE-CHANGE-END

		panic(plainError("send on closed channel"))
	}

	// ADVOCATE-CHANGE-START
	if sg := c.recvq.dequeue(replayElem); sg != nil {
		if !ignored && !c.advocateIgnore {
			AdvocateChanPost(advocateIndex)
		}
		// ADVOCATE-CHANGE-END
		// Found a waiting receiver. We pass the value we want to send
		// directly to the receiver, bypassing the channel buffer (if any).
		send(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true
	}

	if c.qcount < c.dataqsiz {
		// ADVOCATE-CHANGE-START
		if !ignored && !c.advocateIgnore {
			AdvocateChanPost(advocateIndex)
		}
		// ADVOCATE-CHANGE-END
		// Space is available in the channel buffer. Enqueue the element to send.
		qp := chanbuf(c, c.sendx)
		if raceenabled {
			racenotify(c, c.sendx, nil)
		}
		typedmemmove(c.elemtype, qp, ep)
		c.sendx++
		if c.sendx == c.dataqsiz {
			c.sendx = 0
		}
		c.qcount++
		unlock(&c.lock)
		return true
	}

	if !block {
		// ADVOCATE-CHANGE-START
		if !ignored && !c.advocateIgnore {
			AdvocateChanPost(advocateIndex)
		}
		// ADVOCATE-CHANGE-END
		unlock(&c.lock)
		return false
	}

	// Block on the channel. Some receiver will complete our operation for us.
	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	// No stack splits between assigning elem and enqueuing mysg
	// on gp.waiting where copystack can find it.
	mysg.elem = ep
	mysg.waitlink = nil
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	// ADVOCATE-CHANGE-START
	// save partner file and line in sudog
	if replayEnabled && !ignored && !c.advocateIgnore {
		mysg.replayEnabled = true
		mysg.pFile = replayElem.PFile
		mysg.pLine = replayElem.PLine
	}
	// ADVOCATE-CHANGE-END
	gp.waiting = mysg
	gp.param = nil
	c.sendq.enqueue(mysg)
	// Signal to anyone trying to shrink our stack that we're about
	// to park on a channel. The window between when this G's status
	// changes and when we set gp.activeStackChans is not safe for
	// stack shrinking.
	gp.parkingOnChan.Store(true)
	// ADVOCATE-NOTE-START
	// gopark blocks the routine if no communication partner is available
	// and the has no free buffe.
	// ADVOCATE-NOTE-END
	gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceBlockChanSend, 2)
	// Ensure the value being sent is kept alive until the
	// receiver copies it out. The sudog has a pointer to the
	// stack object, but sudogs aren't considered as roots of the
	// stack tracer.
	KeepAlive(ep)

	// someone woke us up.
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	// ADVOCATE-CHANGE-START
	if !ignored && !c.advocateIgnore {
		AdvocateChanPost(advocateIndex)
	}
	// ADVOCATE-CHANGE-END
	gp.waiting = nil
	gp.activeStackChans = false
	closed := !mysg.success
	gp.param = nil
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	mysg.c = nil
	releaseSudog(mysg)
	if closed {
		// ADVOCATE-CHANGE-START
		if !ignored && !c.advocateIgnore {
			AdvocateChanPostCausedByClose(advocateIndex)
		}
		// ADVOCATE-CHANGE-END
		if c.closed == 0 {
			throw("chansend: spurious wakeup")
		}
		if enabled {
			IsNextElementReplayEnd(ExitCodeSendClose, true, false)
		}

		// ADVOCATE-CHANGE-START
		if !ignored && !c.advocateIgnore {
			AdvocateChanPostCausedByClose(advocateIndex)
		}
		// ADVOCATE-CHANGE-END

		panic(plainError("send on closed channel"))
	}
	return true
}

// send processes a send operation on an empty channel c.
// The value ep sent by the sender is copied to the receiver sg.
// The receiver is then woken up to go on its merry way.
// Channel c must be empty and locked.  send unlocks c with unlockf.
// sg must already be dequeued from c.
// ep must be non-nil and point to the heap or the caller's stack.
func send(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
	if raceenabled {
		if c.dataqsiz == 0 {
			racesync(c, sg)
		} else {
			// Pretend we go through the buffer, even though
			// we copy directly. Note that we need to increment
			// the head/tail locations only when raceenabled.
			racenotify(c, c.recvx, nil)
			racenotify(c, c.recvx, sg)
			c.recvx++
			if c.recvx == c.dataqsiz {
				c.recvx = 0
			}
			c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
		}
	}
	if sg.elem != nil {
		sendDirect(c.elemtype, sg, ep)
		sg.elem = nil
	}
	gp := sg.g
	unlockf()
	gp.param = unsafe.Pointer(sg)
	sg.success = true
	if sg.releasetime != 0 {
		sg.releasetime = cputicks()
	}
	goready(gp, skip+1)
}

// Sends and receives on unbuffered or empty-buffered channels are the
// only operations where one running goroutine writes to the stack of
// another running goroutine. The GC assumes that stack writes only
// happen when the goroutine is running and are only done by that
// goroutine. Using a write barrier is sufficient to make up for
// violating that assumption, but the write barrier has to work.
// typedmemmove will call bulkBarrierPreWrite, but the target bytes
// are not in the heap, so that will not help. We arrange to call
// memmove and typeBitsBulkBarrier instead.

func sendDirect(t *_type, sg *sudog, src unsafe.Pointer) {
	// src is on our stack, dst is a slot on another stack.

	// Once we read sg.elem out of sg, it will no longer
	// be updated if the destination's stack gets copied (shrunk).
	// So make sure that no preemption points can happen between read & use.
	dst := sg.elem
	typeBitsBulkBarrier(t, uintptr(dst), uintptr(src), t.Size_)
	// No need for cgo write barrier checks because dst is always
	// Go memory.
	memmove(dst, src, t.Size_)
}

func recvDirect(t *_type, sg *sudog, dst unsafe.Pointer) {
	// dst is on our stack or the heap, src is on another stack.
	// The channel is locked, so src will not move during this
	// operation.
	src := sg.elem
	typeBitsBulkBarrier(t, uintptr(dst), uintptr(src), t.Size_)
	memmove(dst, src, t.Size_)
}

func closechan(c *hchan) {
	if c == nil {
		panic(plainError("close of nil channel"))
	}

	lock(&c.lock)

	// ADVOCATE-CHANGE-START
	// AdvocateChanClose is called when a channel is closed. It creates a close event
	// in the trace.
	if !c.advocateIgnore {
		_, _, _ = WaitForReplay(OperationChannelClose, 2)
		AdvocateChanClose(c.id, c.dataqsiz)
	}
	// ADVOCATE-CHANGE-END

	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("close of closed channel"))
	}

	if raceenabled {
		callerpc := getcallerpc()
		racewritepc(c.raceaddr(), callerpc, abi.FuncPCABIInternal(closechan))
		racerelease(c.raceaddr())
	}

	c.closed = 1

	var glist gList

	// release all readers
	for {
		// ADVOCATE-CHANGE-START
		sg := c.recvq.dequeue(ReplayElement{})
		// ADVOCATE-CHANGE-END
		if sg == nil {
			break
		}
		if sg.elem != nil {
			typedmemclr(c.elemtype, sg.elem)
			sg.elem = nil
		}
		if sg.releasetime != 0 {
			sg.releasetime = cputicks()
		}
		gp := sg.g
		gp.param = unsafe.Pointer(sg)
		sg.success = false
		if raceenabled {
			raceacquireg(gp, c.raceaddr())
		}
		glist.push(gp)
	}

	// release all writers (they will panic)
	for {
		// ADVOCATE-CHANGE-START
		sg := c.sendq.dequeue(ReplayElement{})
		// ADVOCATE-CHANGE-END
		if sg == nil {
			break
		}
		sg.elem = nil
		if sg.releasetime != 0 {
			sg.releasetime = cputicks()
		}
		gp := sg.g
		gp.param = unsafe.Pointer(sg)
		sg.success = false
		if raceenabled {
			raceacquireg(gp, c.raceaddr())
		}
		glist.push(gp)
	}
	unlock(&c.lock)

	// Ready all Gs now that we've dropped the channel lock.
	for !glist.empty() {
		gp := glist.pop()
		gp.schedlink = 0
		goready(gp, 3)
	}
}

// empty reports whether a read from c would block (that is, the channel is
// empty).  It uses a single atomic read of mutable state.
func empty(c *hchan) bool {
	// c.dataqsiz is immutable.
	if c.dataqsiz == 0 {
		return atomic.Loadp(unsafe.Pointer(&c.sendq.first)) == nil
	}
	return atomic.Loaduint(&c.qcount) == 0
}

// entry points for <- c from compiled code.
//
//go:nosplit
func chanrecv1(c *hchan, elem unsafe.Pointer) {
	chanrecv(c, elem, true, false)
}

//go:nosplit
func chanrecv2(c *hchan, elem unsafe.Pointer) (received bool) {
	_, received = chanrecv(c, elem, true, false)
	return
}

// chanrecv receives on channel c and writes the received data to ep.
// ep may be nil, in which case received data is ignored.
// If block == false and no elements are available, returns (false, false).
// Otherwise, if c is closed, zeros *ep and returns (true, false).
// Otherwise, fills in *ep with an element and returns (true, true).
// A non-nil ep must point to the heap or the caller's stack.
// ADVOCATE-CHANGE-START
// set ignored to true, if it is used in a one case + default select. In this case, it is recorded and replayed in the select
func chanrecv(c *hchan, ep unsafe.Pointer, block bool, ignored bool) (selected, received bool) {
	// ADVOCATE-CHANGE-END
	// raceenabled: don't need to check ep, as it is always on the stack
	// or is new memory allocated by reflect.

	if debugChan {
		print("chanrecv: chan=", c, "\n")
	}

	if c == nil {
		if !ignored {
			AdvocateChanRecvPre(0, 0, 0, true)
		}
		if !block {
			return
		}
		gopark(nil, nil, waitReasonChanReceiveNilChan, traceBlockForever, 2)
		throw("unreachable")
	}

	// ADVOCATE-CHANGE-START
	// wait until the replay has reached the current point
	var replayElem ReplayElement
	var enabled bool
	var valid bool
	if !ignored && !c.advocateIgnore {
		enabled, valid, replayElem = WaitForReplay(OperationChannelRecv, 3)
		if enabled && valid {
			if replayElem.Blocked {
				lock(&c.numberRecvMutex)
				c.numberRecv++
				unlock(&c.numberRecvMutex)
				_ = AdvocateChanRecvPre(c.id, c.numberRecv, c.dataqsiz, false)
				BlockForever()
			}
		}
	}
	// ADVOCATE-CHANGE-END

	// Fast path: check for failed non-blocking operation without acquiring the lock.
	if !block && empty(c) {
		// After observing that the channel is not ready for receiving, we observe whether the
		// channel is closed.
		//
		// Reordering of these checks could lead to incorrect behavior when racing with a close.
		// For example, if the channel was open and not empty, was closed, and then drained,
		// reordered reads could incorrectly indicate "open and empty". To prevent reordering,
		// we use atomic loads for both checks, and rely on emptying and closing to happen in
		// separate critical sections under the same lock.  This assumption fails when closing
		// an unbuffered channel with a blocked send, but that is an error condition anyway.
		if atomic.Load(&c.closed) == 0 {
			// Because a channel cannot be reopened, the later observation of the channel
			// being not closed implies that it was also not closed at the moment of the
			// first observation. We behave as if we observed the channel at that moment
			// and report that th,e receive cannot proceed.
			return
		}
		// The channel is irreversibly closed. Re-check whether the channel has any pending data
		// to receive, which could have arrived between the empty and closed checks above.
		// Sequential consistency is also required here, when racing with such a send.
		if empty(c) {
			// The channel is irreversibly closed and empty.
			if raceenabled {
				raceacquire(c.raceaddr())
			}
			if ep != nil {
				typedmemclr(c.elemtype, ep)
			}
			return true, false
		}
	}

	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
	}

	lock(&c.lock)

	// ADVOCATE-CHANGE-START
	// this block is called if a receive is made on a channel.
	// It increases the number of receives on the channel, which is used to
	// identify the communication partner in the advocate analysis.
	// After that a channel receive event is created in the trace to show,
	// that the channel tried to receive.
	// The current function 'chanrecv' only returns, if the receive was successful,
	// meaning the channel either communicated with a send or read from the
	// channel buffer. Therefor, the recive event is modified to include the
	// post information by AdvocateChanPost, if 'chansend' returns.
	// advocateIndex is used to connect the post event to the correct
	// pre envent in the trace.
	var advocateIndex int
	if !ignored && !c.advocateIgnore {
		lock(&c.numberRecvMutex)
		c.numberRecv++
		unlock(&c.numberRecvMutex)
		advocateIndex = AdvocateChanRecvPre(c.id, c.numberRecv, c.dataqsiz, false)
		defer AdvocateChanPost(advocateIndex)
	}
	// ADVOCATE-CHANGE-END

	if c.closed != 0 {
		if c.qcount == 0 {
			if raceenabled {
				raceacquire(c.raceaddr())
			}
			unlock(&c.lock)
			if ep != nil {
				typedmemclr(c.elemtype, ep)
			}
			// ADVOCATE-CHANGE-START
			if !ignored && !c.advocateIgnore {
				AdvocateChanPostCausedByClose(advocateIndex)
			}
			// ADVOCATE-CHANGE-END
			return true, false
		}
		// The channel has been closed, but the channel's buffer have data.
	} else {
		// Just found waiting sender with not closed.
		// ADVOCATE-CHANGE-START
		if sg := c.sendq.dequeue(replayElem); sg != nil {
			// ADVOCATE-CHANGE-END
			// Found a waiting sender. If buffer is size 0, receive value
			// directly from sender. Otherwise, receive from head of queue
			// and add sender's value to the tail of the queue (both map to
			// the same buffer slot because the queue is full).
			recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
			return true, true
		}
	}

	if c.qcount > 0 {
		// Receive directly from queue
		qp := chanbuf(c, c.recvx)
		if raceenabled {
			racenotify(c, c.recvx, nil)
		}
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		typedmemclr(c.elemtype, qp)
		c.recvx++
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		c.qcount--
		unlock(&c.lock)
		return true, true
	}

	if !block {
		unlock(&c.lock)
		return false, false
	}

	// no sender available: block on this channel.
	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	// No stack splits between assigning elem and enqueuing mysg
	// on gp.waiting where copystack can find it.
	mysg.elem = ep
	mysg.waitlink = nil
	gp.waiting = mysg
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	// ADVOCATE-CHANGE-START
	// save partner file and line in sudog
	if replayEnabled && !ignored && !c.advocateIgnore {
		mysg.replayEnabled = true
		mysg.pFile = replayElem.PFile
		mysg.pLine = replayElem.PLine
	}
	// ADVOCATE-CHANGE-END
	gp.param = nil
	c.recvq.enqueue(mysg)
	// Signal to anyone trying to shrink our stack that we're about
	// to park on a channel. The window between when this G's status
	// changes and when we set gp.activeStackChans is not safe for
	// stack shrinking.
	gp.parkingOnChan.Store(true)

	// ADVOCATE-NOTE-START
	// gopark blocks the routine if no communication partner is available
	// and the has no free buffe.
	// ADVOCATE-NOTE-END

	gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanReceive, traceBlockChanRecv, 2)

	// someone woke us up
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	gp.activeStackChans = false
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	success := mysg.success

	// ADVOCATE-CHANGE-START
	if !success && !ignored && !c.advocateIgnore {
		AdvocateChanPostCausedByClose(advocateIndex)
	}
	// ADVOCATE-CHANGE-END

	gp.param = nil
	mysg.c = nil
	releaseSudog(mysg)
	return true, success
}

// recv processes a receive operation on a full channel c.
// There are 2 parts:
//  1. The value sent by the sender sg is put into the channel
//     and the sender is woken up to go on its merry way.
//  2. The value received by the receiver (the current G) is
//     written to ep.
//
// For synchronous channels, both values are the same.
// For asynchronous channels, the receiver gets its data from
// the channel buffer and the sender's data is put in the
// channel buffer.
// Channel c must be full and locked. recv unlocks c with unlockf.
// sg must already be dequeued from c.
// A non-nil ep must point to the heap or the caller's stack.
func recv(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
	if c.dataqsiz == 0 {
		if raceenabled {
			racesync(c, sg)
		}
		if ep != nil {
			// copy data from sender
			recvDirect(c.elemtype, sg, ep)
		}
	} else {
		// Queue is full. Take the item at the
		// head of the queue. Make the sender enqueue
		// its item at the tail of the queue. Since the
		// queue is full, those are both the same slot.
		qp := chanbuf(c, c.recvx)
		if raceenabled {
			racenotify(c, c.recvx, nil)
			racenotify(c, c.recvx, sg)
		}
		// copy data from queue to receiver
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		// copy data from sender to queue
		typedmemmove(c.elemtype, qp, sg.elem)
		c.recvx++
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
	}
	sg.elem = nil
	gp := sg.g
	unlockf()
	gp.param = unsafe.Pointer(sg)
	sg.success = true
	if sg.releasetime != 0 {
		sg.releasetime = cputicks()
	}
	goready(gp, skip+1)
}

func chanparkcommit(gp *g, chanLock unsafe.Pointer) bool {
	// There are unlocked sudogs that point into gp's stack. Stack
	// copying must lock the channels of those sudogs.
	// Set activeStackChans here instead of before we try parking
	// because we could self-deadlock in stack growth on the
	// channel lock.
	gp.activeStackChans = true
	// Mark that it's safe for stack shrinking to occur now,
	// because any thread acquiring this G's stack for shrinking
	// is guaranteed to observe activeStackChans after this store.
	gp.parkingOnChan.Store(false)
	// Make sure we unlock after setting activeStackChans and
	// unsetting parkingOnChan. The moment we unlock chanLock
	// we risk gp getting readied by a channel operation and
	// so gp could continue running before everything before
	// the unlock is visible (even to gp itself).
	unlock((*mutex)(chanLock))
	return true
}

// compiler implements
//
//	select {
//	case c <- v:
//		... foo
//	default:
//		... bar
//	}
//
// as
//
//	if selectnbsend(c, v) {
//		... foo
//	} else {
//		... bar
//	}
func selectnbsend(c *hchan, elem unsafe.Pointer) (selected bool) {
	// ADVOCATE-CHANGE-START
	var replayElem ReplayElement
	var enabled bool
	var valid bool
	if c != nil && !c.advocateIgnore {
		if c != nil && !c.advocateIgnore {
			enabled, valid, replayElem = WaitForReplay(OperationSelect, 2)
			if enabled && valid {
				if replayElem.Blocked {
					lock(&c.numberSendMutex)
					c.numberSend++
					unlock(&c.numberSendMutex)
					_ = AdvocateChanSendPre(c.id, c.numberSend, c.dataqsiz, false)
					BlockForever()
				}
			}
		}
	} else {
		enabled, valid, replayElem = WaitForReplay(OperationSelect, 2)
		if enabled && valid {
			if replayElem.Blocked {
				lock(&c.numberSendMutex)
				c.numberSend++
				unlock(&c.numberSendMutex)
				_ = AdvocateChanSendPre(c.id, c.numberSend, c.dataqsiz, false)
				BlockForever()
			}
		}
	}

	advocateIndex := AdvocateSelectPreOneNonDef(c, true)
	res := chansend(c, elem, false, getcallerpc(), true)
	if c != nil {
		lock(&c.numberSendMutex)
		defer unlock(&c.numberSendMutex)
	}
	AdvocateSelectPostOneNonDef(advocateIndex, res, c)

	return res
	// ADVOCATE-CHANGE-END
}

// compiler implements
//
//	select {
//	case v, ok = <-c:
//		... foo
//	default:
//		... bar
//	}
//
// as
//
//	if selected, ok = selectnbrecv(&v, c); selected {
//		... foo
//	} else {
//		... bar
//	}
func selectnbrecv(elem unsafe.Pointer, c *hchan) (selected, received bool) {
	// ADVOCATE-CHANGE-START
	// see selectnbsend
	var replayElem ReplayElement
	var enabled bool
	var valid bool
	if c != nil && !c.advocateIgnore {
		if !c.advocateIgnore {
			enabled, valid, replayElem = WaitForReplay(OperationSelect, 2)
			if enabled && valid {
				if replayElem.Blocked {
					lock(&c.numberSendMutex)
					c.numberSend++
					unlock(&c.numberSendMutex)
					_ = AdvocateSelectPreOneNonDef(c, false)
					BlockForever()
				}
			}
		}
	} else {
		enabled, valid, replayElem = WaitForReplay(OperationSelect, 2)
		if enabled && valid {
			if replayElem.Blocked {
				lock(&c.numberSendMutex)
				c.numberSend++
				unlock(&c.numberSendMutex)
				_ = AdvocateSelectPreOneNonDef(c, false)
				BlockForever()
			}
		}
	}
	advocateIndex := AdvocateSelectPreOneNonDef(c, false)
	res, recv := chanrecv(c, elem, false, true)
	if c != nil {
		lock(&c.numberRecvMutex)
		defer unlock(&c.numberRecvMutex)
	}
	AdvocateSelectPostOneNonDef(advocateIndex, res, c)
	return res, recv

	// ADVOCATE-CHANGE-END
}

//go:linkname reflect_chansend reflect.chansend0
func reflect_chansend(c *hchan, elem unsafe.Pointer, nb bool) (selected bool) {
	// ADVOCATE-CHANGE-START
	return chansend(c, elem, !nb, getcallerpc(), false)
	// ADVOCATE-CHANGE-END
}

//go:linkname reflect_chanrecv reflect.chanrecv
func reflect_chanrecv(c *hchan, nb bool, elem unsafe.Pointer) (selected bool, received bool) {
	// ADVOCATE-CHANGE-START
	return chanrecv(c, elem, !nb, false)
	// ADVOCATE-CHANGE-END
}

//go:linkname reflect_chanlen reflect.chanlen
func reflect_chanlen(c *hchan) int {
	if c == nil {
		return 0
	}
	return int(c.qcount)
}

//go:linkname reflectlite_chanlen internal/reflectlite.chanlen
func reflectlite_chanlen(c *hchan) int {
	if c == nil {
		return 0
	}
	return int(c.qcount)
}

//go:linkname reflect_chancap reflect.chancap
func reflect_chancap(c *hchan) int {
	if c == nil {
		return 0
	}
	return int(c.dataqsiz)
}

//go:linkname reflect_chanclose reflect.chanclose
func reflect_chanclose(c *hchan) {
	closechan(c)
}

func (q *waitq) enqueue(sgp *sudog) {
	sgp.next = nil
	x := q.last
	if x == nil {
		sgp.prev = nil
		q.first = sgp
		q.last = sgp
		return
	}
	sgp.prev = x
	x.next = sgp
	q.last = sgp
}

// ADVOCATE-CHANGE-START
func (q *waitq) dequeue(rElem ReplayElement) *sudog {
	// ADVOCATE-CHANGE-END
	for {
		sgp := q.first
		if sgp == nil {
			return nil
		}

		// ADVOCATE-CHANGE-START
		// if the channel partner is not correct, the goroutine is not woken up
		// TODO: durch die ganze Queue durchgehen, ob partner vorhanden, um
		// zu verhindern, das Umordnung zu Block führt
		// TODO: oder ganz raus schmeißen, wenn nicht notwendig
		if replayEnabled && !sgp.replayEnabled {
			if !(rElem.File == "") && !sgp.c.advocateIgnore {
				if sgp.pFile != rElem.File || sgp.pLine != rElem.Line {
					return nil
				}
			}
		}
		// ADVOCATE-CHANE-END

		y := sgp.next
		if y == nil {
			q.first = nil
			q.last = nil
		} else {
			y.prev = nil
			q.first = y
			sgp.next = nil // mark as removed (see dequeueSudoG)
		}

		// if a goroutine was put on this queue because of a
		// select, there is a small window between the goroutine
		// being woken up by a different case and it grabbing the
		// channel locks. Once it has the lock
		// it removes itself from the queue, so we won't see it after that.
		// We use a flag in the G struct to tell us when someone
		// else has won the race to signal this goroutine but the goroutine
		// hasn't removed itself from the queue yet.
		if sgp.isSelect && !sgp.g.selectDone.CompareAndSwap(0, 1) {
			continue
		}

		return sgp
	}
}

func (c *hchan) raceaddr() unsafe.Pointer {
	// Treat read-like and write-like operations on the channel to
	// happen at this address. Avoid using the address of qcount
	// or dataqsiz, because the len() and cap() builtins read
	// those addresses, and we don't want them racing with
	// operations like close().
	return unsafe.Pointer(&c.buf)
}

func racesync(c *hchan, sg *sudog) {
	racerelease(chanbuf(c, 0))
	raceacquireg(sg.g, chanbuf(c, 0))
	racereleaseg(sg.g, chanbuf(c, 0))
	raceacquire(chanbuf(c, 0))
}

// Notify the race detector of a send or receive involving buffer entry idx
// and a channel c or its communicating partner sg.
// This function handles the special case of c.elemsize==0.
func racenotify(c *hchan, idx uint, sg *sudog) {
	// We could have passed the unsafe.Pointer corresponding to entry idx
	// instead of idx itself.  However, in a future version of this function,
	// we can use idx to better handle the case of elemsize==0.
	// A future improvement to the detector is to call TSan with c and idx:
	// this way, Go will continue to not allocating buffer entries for channels
	// of elemsize==0, yet the race detector can be made to handle multiple
	// sync objects underneath the hood (one sync object per idx)
	qp := chanbuf(c, idx)
	// When elemsize==0, we don't allocate a full buffer for the channel.
	// Instead of individual buffer entries, the race detector uses the
	// c.buf as the only buffer entry.  This simplification prevents us from
	// following the memory model's happens-before rules (rules that are
	// implemented in racereleaseacquire).  Instead, we accumulate happens-before
	// information in the synchronization object associated with c.buf.
	if c.elemsize == 0 {
		if sg == nil {
			raceacquire(qp)
			racerelease(qp)
		} else {
			raceacquireg(sg.g, qp)
			racereleaseg(sg.g, qp)
		}
	} else {
		if sg == nil {
			racereleaseacquire(qp)
		} else {
			racereleaseacquireg(sg.g, qp)
		}
	}
}
