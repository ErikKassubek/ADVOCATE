// File: goCR_atomic.go
// Brief: Copies to use from the sync/atomic/type.go
// 	They are identical to the others except for the WaitForReplay skip counter
//
// Created: 2024-02-16
//
// License: BSD-3-Clause

package atomic

import "runtime"

// Copies to use from the sync/atomic/type.go
// They are identical to the others except for the WaitForReplay skip counter

func SwapInt32GoCRType(addr *int32, new int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapInt32GoCR(addr, new)
}

func SwapInt64GoCRType(addr *int64, new int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapInt64GoCR(addr, new)
}

func SwapUint32GoCRType(addr *uint32, new uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapUint32GoCR(addr, new)
}

func SwapUint64GoCRType(addr *uint64, new uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapUint64GoCR(addr, new)
}

func SwapUintptrGoCRType(addr *uintptr, new uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapUintptrGoCR(addr, new)
}

func CompareAndSwapInt32GoCRType(addr *int32, old, new int32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapInt32GoCR(addr, old, new)
}

func CompareAndSwapInt64GoCRType(addr *int64, old, new int64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapInt64GoCR(addr, old, new)
}

func CompareAndSwapUint32GoCRType(addr *uint32, old, new uint32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapUint32GoCR(addr, old, new)
}

func CompareAndSwapUint64GoCRType(addr *uint64, old, new uint64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapUint64GoCR(addr, old, new)
}

func CompareAndSwapUintptrGoCRType(addr *uintptr, old, new uintptr) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapUintptrGoCR(addr, old, new)
}

func AddInt32GoCRType(addr *int32, delta int32) (new int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddInt32GoCR(addr, delta)
}

func AddUint32GoCRType(addr *uint32, delta uint32) (new uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddUint32GoCR(addr, delta)
}

func AddInt64GoCRType(addr *int64, delta int64) (new int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddInt64GoCR(addr, delta)
}

func AddUint64GoCRType(addr *uint64, delta uint64) (new uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddUint64GoCR(addr, delta)
}

func AddUintptrGoCRType(addr *uintptr, delta uintptr) (new uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddUintptrGoCR(addr, delta)
}

func LoadInt32GoCRType(addr *int32) (val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadInt32GoCR(addr)
}

func LoadInt64GoCRType(addr *int64) (val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadInt64GoCR(addr)
}

func LoadUint32GoCRType(addr *uint32) (val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadUint32GoCR(addr)
}

func LoadUint64GoCRType(addr *uint64) (val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadUint64GoCR(addr)
}

func LoadUintptrGoCRType(addr *uintptr) (val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadUintptrGoCR(addr)
}

func StoreInt32GoCRType(addr *int32, val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreInt32GoCR(addr, val)
}

func StoreInt64GoCRType(addr *int64, val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreInt64GoCR(addr, val)
}

func StoreUint32GoCRType(addr *uint32, val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreUint32GoCR(addr, val)
}

func StoreUint64GoCRType(addr *uint64, val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreUint64GoCR(addr, val)
}

func StoreUintptrGoCRType(addr *uintptr, val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreUintptrGoCR(addr, val)
}

func AndInt64GoCRType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndInt64GoCR(addr, mask)
}

func AndUint64GoCRType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndUint64GoCR(addr, mask)
}

func AndInt32GoCRType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndInt32GoCR(addr, mask)
}

func AndUint32GoCRType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndUint32GoCR(addr, mask)
}

func AndUintptrGoCRType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndUintptrGoCR(addr, mask)
}

func OrInt64GoCRType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrInt64GoCR(addr, mask)
}

func OrUint64GoCRType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrUint64GoCR(addr, mask)
}

func OrInt32GoCRType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrInt32GoCR(addr, mask)
}

func OrUint32GoCRType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrUint32GoCR(addr, mask)
}

func OrUintptrGoCRType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrUintptrGoCR(addr, mask)
}
