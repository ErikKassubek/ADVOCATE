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

func SwapInt32GocdrType(addr *int32, new int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt32Gocdr(addr, new)
}

func SwapInt64GocdrType(addr *int64, new int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt64Gocdr(addr, new)
}

func SwapUint32GocdrType(addr *uint32, new uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint32Gocdr(addr, new)
}

func SwapUint64GocdrType(addr *uint64, new uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint64Gocdr(addr, new)
}

func SwapUintptrGocdrType(addr *uintptr, new uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUintptrGocdr(addr, new)
}

func CompareAndSwapInt32GocdrType(addr *int32, old, new int32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt32Gocdr(addr, old, new)
}

func CompareAndSwapInt64GocdrType(addr *int64, old, new int64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt64Gocdr(addr, old, new)
}

func CompareAndSwapUint32GocdrType(addr *uint32, old, new uint32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint32Gocdr(addr, old, new)
}

func CompareAndSwapUint64GocdrType(addr *uint64, old, new uint64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint64Gocdr(addr, old, new)
}

func CompareAndSwapUintptrGocdrType(addr *uintptr, old, new uintptr) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUintptrGocdr(addr, old, new)
}

func AddInt32GocdrType(addr *int32, delta int32) (new int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt32Gocdr(addr, delta)
}

func AddUint32GocdrType(addr *uint32, delta uint32) (new uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint32Gocdr(addr, delta)
}

func AddInt64GocdrType(addr *int64, delta int64) (new int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt64Gocdr(addr, delta)
}

func AddUint64GocdrType(addr *uint64, delta uint64) (new uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint64Gocdr(addr, delta)
}

func AddUintptrGocdrType(addr *uintptr, delta uintptr) (new uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUintptrGocdr(addr, delta)
}

func LoadInt32GocdrType(addr *int32) (val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt32Gocdr(addr)
}

func LoadInt64GocdrType(addr *int64) (val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt64Gocdr(addr)
}

func LoadUint32GocdrType(addr *uint32) (val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint32Gocdr(addr)
}

func LoadUint64GocdrType(addr *uint64) (val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint64Gocdr(addr)
}

func LoadUintptrGocdrType(addr *uintptr) (val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUintptrGocdr(addr)
}

func StoreInt32GocdrType(addr *int32, val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt32Gocdr(addr, val)
}

func StoreInt64GocdrType(addr *int64, val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt64Gocdr(addr, val)
}

func StoreUint32GocdrType(addr *uint32, val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint32Gocdr(addr, val)
}

func StoreUint64GocdrType(addr *uint64, val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint64Gocdr(addr, val)
}

func StoreUintptrGocdrType(addr *uintptr, val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUintptrGocdr(addr, val)
}

func AndInt64GocdrType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt64Gocdr(addr, mask)
}

func AndUint64GocdrType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint64Gocdr(addr, mask)
}

func AndInt32GocdrType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt32Gocdr(addr, mask)
}

func AndUint32GocdrType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint32Gocdr(addr, mask)
}

func AndUintptrGocdrType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUintptrGocdr(addr, mask)
}

func OrInt64GocdrType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt64Gocdr(addr, mask)
}

func OrUint64GocdrType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint64Gocdr(addr, mask)
}

func OrInt32GocdrType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt32Gocdr(addr, mask)
}

func OrUint32GocdrType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint32Gocdr(addr, mask)
}

func OrUintptrGocdrType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GocdrAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUintptrGocdr(addr, mask)
}
