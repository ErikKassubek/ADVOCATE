// File: goCR_atomic.go
// Brief: Atomic functions with replay and recording
//
// Created: 2024-02-16
//
// License: BSD-3-Clause

package atomic

import (
	"runtime"
	"unsafe"
)

// SwapInt32 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Int32.Swap] instead.
func SwapInt32(addr *int32, new int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 2)
	return SwapInt32GoCR(addr, new)
}

// SwapInt64 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Int64.Swap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func SwapInt64(addr *int64, new int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 2)
	return SwapInt64GoCR(addr, new)
}

// SwapUint32 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Uint32.Swap] instead.
func SwapUint32(addr *uint32, new uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 2)
	return SwapUint32GoCR(addr, new)
}

// SwapUint64 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Uint64.Swap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func SwapUint64(addr *uint64, new uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 2)
	return SwapUint64GoCR(addr, new)
}

// SwapUintptr atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Uintptr.Swap] instead.
func SwapUintptr(addr *uintptr, new uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicSwap, 2)
	return SwapUintptrGoCR(addr, new)
}

// SwapPointer atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Pointer.Swap] instead.
// func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) (old unsafe.Pointer) {
// 	return SwapPointerGoCR(addr, new)
// }

// CompareAndSwapInt32 executes the compare-and-swap operation for an int32 value.
// Consider using the more ergonomic and less error-prone [Int32.CompareAndSwap] instead.
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 2)
	return CompareAndSwapInt32GoCR(addr, old, new)
}

// CompareAndSwapInt64 executes the compare-and-swap operation for an int64 value.
// Consider using the more ergonomic and less error-prone [Int64.CompareAndSwap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 2)
	return CompareAndSwapInt64GoCR(addr, old, new)
}

// CompareAndSwapUint32 executes the compare-and-swap operation for a uint32 value.
// Consider using the more ergonomic and less error-prone [Uint32.CompareAndSwap] instead.
func CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 2)
	return CompareAndSwapUint32GoCR(addr, old, new)
}

// CompareAndSwapUint64 executes the compare-and-swap operation for a uint64 value.
// Consider using the more ergonomic and less error-prone [Uint64.CompareAndSwap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 2)
	return CompareAndSwapUint64GoCR(addr, old, new)
}

// CompareAndSwapUintptr executes the compare-and-swap operation for a uintptr value.
// Consider using the more ergonomic and less error-prone [Uintptr.CompareAndSwap] instead.
func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicCompareAndSwap, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicCompareAndSwap, 2)
	return CompareAndSwapUintptrGoCR(addr, old, new)
}

// // CompareAndSwapPointer executes the compare-and-swap operation for a unsafe.Pointer value.
// // Consider using the more ergonomic and less error-prone [Pointer.CompareAndSwap] instead.
// func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool) {
// 	return CompareAndSwapPointerGoCR(addr, old, new)
// }

// AddInt32 atomically adds delta to *addr and returns the new value.
// Consider using the more ergonomic and less error-prone [Int32.Add] instead.
func AddInt32(addr *int32, delta int32) (new int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 2)
	return AddInt32GoCR(addr, delta)
}

// AddUint32 atomically adds delta to *addr and returns the new value.
// To subtract a signed positive constant value c from x, do AddUint32(&x, ^uint32(c-1)).
// In particular, to decrement x, do AddUint32(&x, ^uint32(0)).
// Consider using the more ergonomic and less error-prone [Uint32.Add] instead.
func AddUint32(addr *uint32, delta uint32) (new uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 2)
	return AddUint32GoCR(addr, delta)
}

// AddInt64 atomically adds delta to *addr and returns the new value.
// Consider using the more ergonomic and less error-prone [Int64.Add] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func AddInt64(addr *int64, delta int64) (new int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 2)
	return AddInt64GoCR(addr, delta)
}

// AddUint64 atomically adds delta to *addr and returns the new value.
// To subtract a signed positive constant value c from x, do AddUint64(&x, ^uint64(c-1)).
// In particular, to decrement x, do AddUint64(&x, ^uint64(0)).
// Consider using the more ergonomic and less error-prone [Uint64.Add] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func AddUint64(addr *uint64, delta uint64) (new uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 2)
	return AddUint64GoCR(addr, delta)
}

// AddUintptr atomically adds delta to *addr and returns the new value.
// Consider using the more ergonomic and less error-prone [Uintptr.Add] instead.
func AddUintptr(addr *uintptr, delta uintptr) (new uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAdd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAdd, 2)
	return AddUintptrGoCR(addr, delta)
}

// LoadInt32 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Int32.Load] instead.
func LoadInt32(addr *int32) (val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 2)
	return LoadInt32GoCR(addr)
}

// LoadInt64 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Int64.Load] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func LoadInt64(addr *int64) (val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 2)
	return LoadInt64GoCR(addr)
}

// LoadUint32 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Uint32.Load] instead.
func LoadUint32(addr *uint32) (val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 2)
	return LoadUint32GoCR(addr)
}

// LoadUint64 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Uint64.Load] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func LoadUint64(addr *uint64) (val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 2)
	return LoadUint64GoCR(addr)
}

// LoadUintptr atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Uintptr.Load] instead.
func LoadUintptr(addr *uintptr) (val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 2)
	return LoadUintptrGoCR(addr)
}

// LoadPointer atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Pointer.Load] instead.
func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicLoad, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicLoad, 2)
	return LoadPointerGoCR(addr)
}

// StoreInt32 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Int32.Store] instead.
func StoreInt32(addr *int32, val int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 2)
	StoreInt32GoCR(addr, val)
}

// StoreInt64 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Int64.Store] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func StoreInt64(addr *int64, val int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 2)
	StoreInt64GoCR(addr, val)
}

// StoreUint32 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Uint32.Store] instead.
func StoreUint32(addr *uint32, val uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 2)
	StoreUint32GoCR(addr, val)
}

// StoreUint64 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Uint64.Store] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func StoreUint64(addr *uint64, val uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 2)
	StoreUint64GoCR(addr, val)
}

// StoreUintptr atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Uintptr.Store] instead.
func StoreUintptr(addr *uintptr, val uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 2)
	StoreUintptrGoCR(addr, val)
}

// StorePointer atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Pointer.Store] instead.
// func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer) {
// 	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicStore, 2, true)
// 	if wait {
// 		defer func() { chAck <- struct{}{} }()
// 		<-chWait
// 	}
// 	runtime.GoCRAtomic(addr, runtime.OperationAtomicStore, 2)
// 	StorePointerGoCR(addr, val)
// }

func AndInt64(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 2)
	return AndInt64GoCR(addr, mask)
}

func AndUint64(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 2)
	return AndUint64GoCR(addr, mask)
}

func AndInt32(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 2)
	return AndInt32GoCR(addr, mask)
}

func AndUint32(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 2)
	return AndUint32GoCR(addr, mask)
}

func AndUintptr(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicAnd, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicAnd, 2)
	return AndUintptrGoCR(addr, mask)
}

func OrInt64(addr *int64, mask int64) (old int64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 2)
	return OrInt64GoCR(addr, mask)
}

func OrUint64(addr *uint64, mask uint64) (old uint64) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 2)
	return OrUint64GoCR(addr, mask)
}

func OrInt32(addr *int32, mask int32) (old int32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 2)
	return OrInt32GoCR(addr, mask)
}

func OrUint32(addr *uint32, mask uint32) (old uint32) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 2)
	return OrUint32GoCR(addr, mask)
}

func OrUintptr(addr *uintptr, mask uintptr) (old uintptr) {
	wait, chWait, chAck, _ := runtime.WaitForReplay(runtime.OperationAtomicOr, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-chWait
	}
	runtime.GoCRAtomic(addr, runtime.OperationAtomicOr, 2)
	return OrUintptrGoCR(addr, mask)
}
