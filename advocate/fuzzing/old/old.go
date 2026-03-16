// // Copyright (c) 2025 Erik Kassubek
// //
// // File: precomputations.go
// // Brief: Calculate all required ress for a trace
// //
// // Author: Erik Kassubek
// // Created: 2025-12-04
// //
// // License: BSD-3-Clause

package equivalence

// import (
// 	"advocate/trace"
// 	"advocate/utils/log"
// 	"advocate/utils/types"
// )

// var (
// // traceID -> ress
// )

// var (
// 	lastWriter = make(map[int]int)
// 	send       = make(map[int]map[int]int)       // channel id -> counter -> traceID
// 	recv       = make(map[int]map[int]int)       // channel id -> counter -> traceID
// 	numberSend = make(map[int]int)               // id -> number send
// 	numberRecv = make(map[int]int)               // id -> number recv/close
// 	mut        = make(map[int]*types.Stack[int]) // mut id -> lock ops
// 	rlock      = make(map[int]struct{})          // trace id
// 	wgCounter  = make(map[int]int)               // obID -> counter after op
// 	onceSuc    = make(map[int]struct{})          // obID
// 	lastRel    = make(map[int]int)               // opjID -> last wg signal/broadcast id
// )

// func precompute(t *TraceMin) map[int]types.Pair[int, int] {
// 	values := make(map[int]types.Pair[int, int])

// 	for _, elem := range t.Trace() {
// 		var res types.Pair[int, int]

// 		switch elem.GetType(false) {
// 		case trace.Atomic:
// 			res = preCompAtomic(&elem)
// 		case trace.Channel, trace.Select:
// 			res = preCompChannel(&elem)
// 		case trace.Mutex:
// 			res = preCompMutex(&elem)
// 		case trace.Wait:
// 			res = preCompWg(&elem)
// 		case trace.Cond:
// 			res = preCompCondVar(&elem)
// 		case trace.Once:
// 			res = preCompOnce(&elem)
// 		}

// 		values[elem.GetID()] = res
// 	}

// 	return values
// }

// // For each element, store the id of the last writer
// func preCompAtomic(elem *ElemMin) types.Pair[int, int] {
// 	res := types.NewPair(0, 0)

// 	if id, ok := lastWriter[elem.ObjID]; ok {
// 		res.X = id
// 	}

// 	t := elem.GetType(true)

// 	// if t is write
// 	if t != trace.AtomicLoad {
// 		lastWriter[elem.ObjID] = elem.ID
// 	}

// 	return res
// }

// // for each channel/select store the id of the comm partner
// // for each select, store the index of the executed operation
// func preCompChannel(elem *trace.ElemMin) types.Pair[int, int] {
// 	t := elem.GetType(true)
// 	objID := elem.ObjID
// 	id := elem.ID

// 	res := types.NewPair(0, 0)

// 	switch t {
// 	case trace.ChannelSend:
// 		numberSend[objID]++
// 		if r, ok := recv[objID][numberSend[objID]]; ok {
// 			res.X = r
// 		}
// 		send[objID][numberSend[objID]] = id
// 	case trace.ChannelRecv, trace.ChannelClose:
// 		numberRecv[objID]++
// 		if r, ok := send[objID][numberRecv[objID]]; ok {
// 			res.X = r
// 		}
// 		recv[objID][numberSend[objID]] = id
// 	case trace.SelectOp:
// 		chosenInd := elem.Value

// 		if chosenInd < 0 || chosenInd >= len(elem.Channel) {
// 			log.Errorf("Invalid chosen index %d", chosenInd)
// 			return res
// 		}

// 		chosenCase := elem.Channel[chosenInd]

// 		chanID := chosenCase.X
// 		if chosenCase.Y { // case is send
// 			numberSend[chanID]++
// 			if r, ok := recv[chanID][numberSend[chanID]]; ok {
// 				res.X = r
// 			}
// 			send[chanID][numberSend[chanID]] = id
// 		} else { // case is recv
// 			numberRecv[chanID]++
// 			if r, ok := send[chanID][numberRecv[chanID]]; ok {
// 				res.X = r
// 			}
// 			recv[chanID][numberSend[chanID]] = id
// 		}

// 		res.Y = chosenInd
// 	}

// 	return res
// }

// // for each mutex unlock, store the corresponding lock
// // for each trylock store 1 if it is successful, 0 if not
// func preCompMutex(elem *trace.ElemMin) types.Pair[int, int] {
// 	objId := elem.ObjID
// 	id := elem.ID
// 	t := elem.GetType(true)

// 	if _, ok := mut[objId]; !ok {
// 		mut[objId] = types.NewStack[int]()
// 	}

// 	res := types.NewPair(0, 0)

// 	switch t {
// 	case trace.MutexUnlock, trace.MutexRUnlock:
// 		res.X = mut[objId].Pop()
// 	case trace.MutexTryLock:
// 		if mut[objId].IsEmpty() {
// 			mut[objId].Push(id)
// 			res.Y = 1
// 		} else {
// 			res.Y = 0
// 		}
// 	case trace.MutexTryRLock:
// 		rlock[id] = struct{}{}
// 		if mut[objId].IsEmpty() {
// 			mut[objId].Push(id)
// 			res.X = 1
// 		} else if _, ok := rlock[mut[objId].Peek()]; ok {
// 			mut[objId].Push(id)
// 			res.X = 1
// 		} else {
// 			res.Y = 0
// 		}
// 	case trace.MutexRLock:
// 		rlock[id] = struct{}{}
// 		mut[objId].Push(id)
// 	case trace.MutexLock:
// 		mut[objId].Push(id)
// 	}

// 	return res
// }

// // set the res to 1 if the expected counter res after this op is > 0,
// // 0 if it is 0 and -1 if it is <0
// func preCompWg(elem *trace.ElemMin) types.Pair[int, int] {
// 	objID := elem.ObjID

// 	wgCounter[objID] += elem.Value

// 	newVal := wgCounter[objID]

// 	res := types.NewPair(0, 0)

// 	if newVal > 0 {
// 		res.X = 1
// 	} else if newVal < 0 {
// 		res.X = -1
// 	}

// 	return res
// }

// // for each wait store the id of the corresponding signal/bradcast
// // todo: not sure if this is always correct. If we get two signals directly
// // next to each other and the first releases the wait, but the second
// // is recorded before the post of the wait, it would be wrong.
// // How likely is this?
// func preCompCondVar(elem *trace.ElemMin) types.Pair[int, int] {
// 	t := elem.GetType(true)
// 	objId := elem.ObjID

// 	res := types.NewPair(0, 0)

// 	switch t {
// 	case trace.CondWait:
// 		if val, ok := lastRel[objId]; ok {
// 			res.X = val
// 		} else {
// 			log.Errorf("Invalid cond var release")
// 		}
// 	case trace.CondSignal, trace.CondBroadcast:
// 		lastRel[objId] = elem.ID
// 	}

// 	return res
// }

// // store 1 if it is the first once on the var, 0 otherwise
// func preCompOnce(elem *trace.ElemMin) types.Pair[int, int] {
// 	res := types.NewPair(0, 0)
// 	objId := elem.ObjID

// 	if _, ok := onceSuc[objId]; ok {
// 		res.X = 1
// 	} else {
// 		onceSuc[objId] = struct{}{}
// 		res.X = 0
// 	}

// 	return res
// }
