// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_atomic.go
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

func SwapInt32AdvocateType(addr *int32, new int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapInt32Advocate(addr, new)
}

func SwapInt64AdvocateType(addr *int64, new int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapInt64Advocate(addr, new)
}

func SwapUint32AdvocateType(addr *uint32, new uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapUint32Advocate(addr, new)
}

func SwapUint64AdvocateType(addr *uint64, new uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapUint64Advocate(addr, new)
}

func SwapUintptrAdvocateType(addr *uintptr, new uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicSwap, 3)
	return SwapUintptrAdvocate(addr, new)
}

func CompareAndSwapInt32AdvocateType(addr *int32, old, new int32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapInt32Advocate(addr, old, new)
}

func CompareAndSwapInt64AdvocateType(addr *int64, old, new int64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapInt64Advocate(addr, old, new)
}

func CompareAndSwapUint32AdvocateType(addr *uint32, old, new uint32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapUint32Advocate(addr, old, new)
}

func CompareAndSwapUint64AdvocateType(addr *uint64, old, new uint64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapUint64Advocate(addr, old, new)
}

func CompareAndSwapUintptrAdvocateType(addr *uintptr, old, new uintptr) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicCompareAndSwap, 3)
	return CompareAndSwapUintptrAdvocate(addr, old, new)
}

func AddInt32AdvocateType(addr *int32, delta int32) (new int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddInt32Advocate(addr, delta)
}

func AddUint32AdvocateType(addr *uint32, delta uint32) (new uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddUint32Advocate(addr, delta)
}

func AddInt64AdvocateType(addr *int64, delta int64) (new int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddInt64Advocate(addr, delta)
}

func AddUint64AdvocateType(addr *uint64, delta uint64) (new uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddUint64Advocate(addr, delta)
}

func AddUintptrAdvocateType(addr *uintptr, delta uintptr) (new uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAdd, 3)
	return AddUintptrAdvocate(addr, delta)
}

func LoadInt32AdvocateType(addr *int32) (val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadInt32Advocate(addr)
}

func LoadInt64AdvocateType(addr *int64) (val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadInt64Advocate(addr)
}

func LoadUint32AdvocateType(addr *uint32) (val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadUint32Advocate(addr)
}

func LoadUint64AdvocateType(addr *uint64) (val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadUint64Advocate(addr)
}

func LoadUintptrAdvocateType(addr *uintptr) (val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicLoad, 3)
	return LoadUintptrAdvocate(addr)
}

func StoreInt32AdvocateType(addr *int32, val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreInt32Advocate(addr, val)
}

func StoreInt64AdvocateType(addr *int64, val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreInt64Advocate(addr, val)
}

func StoreUint32AdvocateType(addr *uint32, val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreUint32Advocate(addr, val)
}

func StoreUint64AdvocateType(addr *uint64, val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreUint64Advocate(addr, val)
}

func StoreUintptrAdvocateType(addr *uintptr, val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicStore, 3)
	StoreUintptrAdvocate(addr, val)
}

func AndInt64AdvocateType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndInt64Advocate(addr, mask)
}

func AndUint64AdvocateType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndUint64Advocate(addr, mask)
}

func AndInt32AdvocateType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndInt32Advocate(addr, mask)
}

func AndUint32AdvocateType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndUint32Advocate(addr, mask)
}

func AndUintptrAdvocateType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicAnd, 3)
	return AndUintptrAdvocate(addr, mask)
}

func OrInt64AdvocateType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrInt64Advocate(addr, mask)
}

func OrUint64AdvocateType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrUint64Advocate(addr, mask)
}

func OrInt32AdvocateType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrInt32Advocate(addr, mask)
}

func OrUint32AdvocateType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrUint32Advocate(addr, mask)
}

func OrUintptrAdvocateType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 3, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.AdvocateAtomic(addr, runtime.OperationAtomicOr, 3)
	return OrUintptrAdvocate(addr, mask)
}
