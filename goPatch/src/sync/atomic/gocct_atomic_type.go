// Copyright (c) 2025 Erik Kassubek
//
// File: gocct_atomic.go
// Brief: Copies to use from the sync/atomic/type.go
// 	They are identical to the others except for the WaitForReplay skip counter
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package atomic

import "runtime"

// Copies to use from the sync/atomic/type.go
// They are identical to the others except for the WaitForReplay skip counter

func SwapInt32GoCCTType(addr *int32, new int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt32GoCCT(addr, new)
}

func SwapInt64GoCCTType(addr *int64, new int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt64GoCCT(addr, new)
}

func SwapUint32GoCCTType(addr *uint32, new uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint32GoCCT(addr, new)
}

func SwapUint64GoCCTType(addr *uint64, new uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint64GoCCT(addr, new)
}

func SwapUintptrGoCCTType(addr *uintptr, new uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUintptrGoCCT(addr, new)
}

func CompareAndSwapInt32GoCCTType(addr *int32, old, new int32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt32GoCCT(addr, old, new)
}

func CompareAndSwapInt64GoCCTType(addr *int64, old, new int64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt64GoCCT(addr, old, new)
}

func CompareAndSwapUint32GoCCTType(addr *uint32, old, new uint32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint32GoCCT(addr, old, new)
}

func CompareAndSwapUint64GoCCTType(addr *uint64, old, new uint64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint64GoCCT(addr, old, new)
}

func CompareAndSwapUintptrGoCCTType(addr *uintptr, old, new uintptr) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUintptrGoCCT(addr, old, new)
}

func AddInt32GoCCTType(addr *int32, delta int32) (new int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt32GoCCT(addr, delta)
}

func AddUint32GoCCTType(addr *uint32, delta uint32) (new uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint32GoCCT(addr, delta)
}

func AddInt64GoCCTType(addr *int64, delta int64) (new int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt64GoCCT(addr, delta)
}

func AddUint64GoCCTType(addr *uint64, delta uint64) (new uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint64GoCCT(addr, delta)
}

func AddUintptrGoCCTType(addr *uintptr, delta uintptr) (new uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUintptrGoCCT(addr, delta)
}

func LoadInt32GoCCTType(addr *int32) (val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt32GoCCT(addr)
}

func LoadInt64GoCCTType(addr *int64) (val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt64GoCCT(addr)
}

func LoadUint32GoCCTType(addr *uint32) (val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint32GoCCT(addr)
}

func LoadUint64GoCCTType(addr *uint64) (val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint64GoCCT(addr)
}

func LoadUintptrGoCCTType(addr *uintptr) (val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUintptrGoCCT(addr)
}

func StoreInt32GoCCTType(addr *int32, val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt32GoCCT(addr, val)
}

func StoreInt64GoCCTType(addr *int64, val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt64GoCCT(addr, val)
}

func StoreUint32GoCCTType(addr *uint32, val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint32GoCCT(addr, val)
}

func StoreUint64GoCCTType(addr *uint64, val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint64GoCCT(addr, val)
}

func StoreUintptrGoCCTType(addr *uintptr, val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUintptrGoCCT(addr, val)
}

func AndInt64GoCCTType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt64GoCCT(addr, mask)
}

func AndUint64GoCCTType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint64GoCCT(addr, mask)
}

func AndInt32GoCCTType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt32GoCCT(addr, mask)
}

func AndUint32GoCCTType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint32GoCCT(addr, mask)
}

func AndUintptrGoCCTType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUintptrGoCCT(addr, mask)
}

func OrInt64GoCCTType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt64GoCCT(addr, mask)
}

func OrUint64GoCCTType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint64GoCCT(addr, mask)
}

func OrInt32GoCCTType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt32GoCCT(addr, mask)
}

func OrUint32GoCCTType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint32GoCCT(addr, mask)
}

func OrUintptrGoCCTType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCCTAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUintptrGoCCT(addr, mask)
}
