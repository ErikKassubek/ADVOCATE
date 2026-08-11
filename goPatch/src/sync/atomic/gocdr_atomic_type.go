// Copyright (c) 2025 Erik Kassubek
//
// File: gocdr_atomic.go
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

func SwapInt32GoCDRType(addr *int32, new int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt32GoCDR(addr, new)
}

func SwapInt64GoCDRType(addr *int64, new int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt64GoCDR(addr, new)
}

func SwapUint32GoCDRType(addr *uint32, new uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint32GoCDR(addr, new)
}

func SwapUint64GoCDRType(addr *uint64, new uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint64GoCDR(addr, new)
}

func SwapUintptrGoCDRType(addr *uintptr, new uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUintptrGoCDR(addr, new)
}

func CompareAndSwapInt32GoCDRType(addr *int32, old, new int32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt32GoCDR(addr, old, new)
}

func CompareAndSwapInt64GoCDRType(addr *int64, old, new int64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt64GoCDR(addr, old, new)
}

func CompareAndSwapUint32GoCDRType(addr *uint32, old, new uint32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint32GoCDR(addr, old, new)
}

func CompareAndSwapUint64GoCDRType(addr *uint64, old, new uint64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint64GoCDR(addr, old, new)
}

func CompareAndSwapUintptrGoCDRType(addr *uintptr, old, new uintptr) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUintptrGoCDR(addr, old, new)
}

func AddInt32GoCDRType(addr *int32, delta int32) (new int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt32GoCDR(addr, delta)
}

func AddUint32GoCDRType(addr *uint32, delta uint32) (new uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint32GoCDR(addr, delta)
}

func AddInt64GoCDRType(addr *int64, delta int64) (new int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt64GoCDR(addr, delta)
}

func AddUint64GoCDRType(addr *uint64, delta uint64) (new uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint64GoCDR(addr, delta)
}

func AddUintptrGoCDRType(addr *uintptr, delta uintptr) (new uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUintptrGoCDR(addr, delta)
}

func LoadInt32GoCDRType(addr *int32) (val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt32GoCDR(addr)
}

func LoadInt64GoCDRType(addr *int64) (val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt64GoCDR(addr)
}

func LoadUint32GoCDRType(addr *uint32) (val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint32GoCDR(addr)
}

func LoadUint64GoCDRType(addr *uint64) (val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint64GoCDR(addr)
}

func LoadUintptrGoCDRType(addr *uintptr) (val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUintptrGoCDR(addr)
}

func StoreInt32GoCDRType(addr *int32, val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt32GoCDR(addr, val)
}

func StoreInt64GoCDRType(addr *int64, val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt64GoCDR(addr, val)
}

func StoreUint32GoCDRType(addr *uint32, val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint32GoCDR(addr, val)
}

func StoreUint64GoCDRType(addr *uint64, val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint64GoCDR(addr, val)
}

func StoreUintptrGoCDRType(addr *uintptr, val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUintptrGoCDR(addr, val)
}

func AndInt64GoCDRType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt64GoCDR(addr, mask)
}

func AndUint64GoCDRType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint64GoCDR(addr, mask)
}

func AndInt32GoCDRType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt32GoCDR(addr, mask)
}

func AndUint32GoCDRType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint32GoCDR(addr, mask)
}

func AndUintptrGoCDRType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUintptrGoCDR(addr, mask)
}

func OrInt64GoCDRType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt64GoCDR(addr, mask)
}

func OrUint64GoCDRType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint64GoCDR(addr, mask)
}

func OrInt32GoCDRType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt32GoCDR(addr, mask)
}

func OrUint32GoCDRType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint32GoCDR(addr, mask)
}

func OrUintptrGoCDRType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCDRAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUintptrGoCDR(addr, mask)
}
