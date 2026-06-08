// Copyright (c) 2025 Erik Kassubek
//
// File: oosc_atomic.go
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

func SwapInt32OoscType(addr *int32, new int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt32Oosc(addr, new)
}

func SwapInt64OoscType(addr *int64, new int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapInt64Oosc(addr, new)
}

func SwapUint32OoscType(addr *uint32, new uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint32Oosc(addr, new)
}

func SwapUint64OoscType(addr *uint64, new uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUint64Oosc(addr, new)
}

func SwapUintptrOoscType(addr *uintptr, new uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicSwap, runtime.CallerSkipAtomicType)
	return SwapUintptrOosc(addr, new)
}

func CompareAndSwapInt32OoscType(addr *int32, old, new int32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt32Oosc(addr, old, new)
}

func CompareAndSwapInt64OoscType(addr *int64, old, new int64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapInt64Oosc(addr, old, new)
}

func CompareAndSwapUint32OoscType(addr *uint32, old, new uint32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint32Oosc(addr, old, new)
}

func CompareAndSwapUint64OoscType(addr *uint64, old, new uint64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUint64Oosc(addr, old, new)
}

func CompareAndSwapUintptrOoscType(addr *uintptr, old, new uintptr) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicCompareAndSwap, runtime.CallerSkipAtomicType)
	return CompareAndSwapUintptrOosc(addr, old, new)
}

func AddInt32OoscType(addr *int32, delta int32) (new int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt32Oosc(addr, delta)
}

func AddUint32OoscType(addr *uint32, delta uint32) (new uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint32Oosc(addr, delta)
}

func AddInt64OoscType(addr *int64, delta int64) (new int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddInt64Oosc(addr, delta)
}

func AddUint64OoscType(addr *uint64, delta uint64) (new uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUint64Oosc(addr, delta)
}

func AddUintptrOoscType(addr *uintptr, delta uintptr) (new uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAdd, runtime.CallerSkipAtomicType)
	return AddUintptrOosc(addr, delta)
}

func LoadInt32OoscType(addr *int32) (val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt32Oosc(addr)
}

func LoadInt64OoscType(addr *int64) (val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadInt64Oosc(addr)
}

func LoadUint32OoscType(addr *uint32) (val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint32Oosc(addr)
}

func LoadUint64OoscType(addr *uint64) (val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUint64Oosc(addr)
}

func LoadUintptrOoscType(addr *uintptr) (val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicLoad, runtime.CallerSkipAtomicType)
	return LoadUintptrOosc(addr)
}

func StoreInt32OoscType(addr *int32, val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt32Oosc(addr, val)
}

func StoreInt64OoscType(addr *int64, val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreInt64Oosc(addr, val)
}

func StoreUint32OoscType(addr *uint32, val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint32Oosc(addr, val)
}

func StoreUint64OoscType(addr *uint64, val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUint64Oosc(addr, val)
}

func StoreUintptrOoscType(addr *uintptr, val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicStore, runtime.CallerSkipAtomicType)
	StoreUintptrOosc(addr, val)
}

func AndInt64OoscType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt64Oosc(addr, mask)
}

func AndUint64OoscType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint64Oosc(addr, mask)
}

func AndInt32OoscType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndInt32Oosc(addr, mask)
}

func AndUint32OoscType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUint32Oosc(addr, mask)
}

func AndUintptrOoscType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicAnd, runtime.CallerSkipAtomicType)
	return AndUintptrOosc(addr, mask)
}

func OrInt64OoscType(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt64Oosc(addr, mask)
}

func OrUint64OoscType(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint64Oosc(addr, mask)
}

func OrInt32OoscType(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrInt32Oosc(addr, mask)
}

func OrUint32OoscType(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUint32Oosc(addr, mask)
}

func OrUintptrOoscType(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, runtime.CallerSkipAtomicType, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.OoscAtomic(addr, runtime.OperationAtomicOr, runtime.CallerSkipAtomicType)
	return OrUintptrOosc(addr, mask)
}
