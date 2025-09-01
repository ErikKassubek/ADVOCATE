// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || arm || mips || mipsle

// ADVOCATE-START

package atomic

// SwapInt64 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Int64.Swap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func SwapInt64Advocate(addr *int64, new int64) (old int64)

// SwapUint64 atomically stores new into *addr and returns the previous *addr value.
// Consider using the more ergonomic and less error-prone [Uint64.Swap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func SwapUint64Advocate(addr *uint64, new uint64) (old uint64)

// CompareAndSwapInt64 executes the compare-and-swap operation for an int64 value.
// Consider using the more ergonomic and less error-prone [Int64.CompareAndSwap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func CompareAndSwapInt64Advocate(addr *int64, old, new int64) (swapped bool)

// CompareAndSwapUint64 executes the compare-and-swap operation for a uint64 value.
// Consider using the more ergonomic and less error-prone [Uint64.CompareAndSwap] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func CompareAndSwapUint64Advocate(addr *uint64, old, new uint64) (swapped bool)

// AddInt64 atomically adds delta to *addr and returns the new value.
// Consider using the more ergonomic and less error-prone [Int64.Add] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func AddInt64Advocate(addr *int64, delta int64) (new int64)

// AddUint64 atomically adds delta to *addr and returns the new value.
// To subtract a signed positive constant value c from x, do AddUint64(&x, ^uint64(c-1)).
// In particular, to decrement x, do AddUint64(&x, ^uint64(0)).
// Consider using the more ergonomic and less error-prone [Uint64.Add] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func AddUint64Advocate(addr *uint64, delta uint64) (new uint64)

// AndInt64 atomically performs a bitwise AND operation on *addr using the bitmask provided as mask
// and returns the old value.
// Consider using the more ergonomic and less error-prone [Int64.And] instead.
func AndInt64Advocate(addr *int64, mask int64) (old int64)

// AndUint64 atomically performs a bitwise AND operation on *addr using the bitmask provided as mask
// and returns the old.
// Consider using the more ergonomic and less error-prone [Uint64.And] instead.
func AndUint64Advocate(addr *uint64, mask uint64) (old uint64)

// OrInt64 atomically performs a bitwise OR operation on *addr using the bitmask provided as mask
// and returns the old value.
// Consider using the more ergonomic and less error-prone [Int64.Or] instead.
func OrInt64Advocate(addr *int64, mask int64) (old int64)

// OrUint64 atomically performs a bitwise OR operation on *addr using the bitmask provided as mask
// and returns the old value.
// Consider using the more ergonomic and less error-prone [Uint64.Or] instead.
func OrUint64Advocate(addr *uint64, mask uint64) (old uint64)

// LoadInt64 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Int64.Load] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func LoadInt64Advocate(addr *int64) (val int64)

// LoadUint64 atomically loads *addr.
// Consider using the more ergonomic and less error-prone [Uint64.Load] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func LoadUint64Advocate(addr *uint64) (val uint64)

// StoreInt64 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Int64.Store] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func StoreInt64Advocate(addr *int64, val int64)

// StoreUint64 atomically stores val into *addr.
// Consider using the more ergonomic and less error-prone [Uint64.Store] instead
// (particularly if you target 32-bit platforms; see the bugs section).
func StoreUint64Advocate(addr *uint64, val uint64)

// ADVOCATE-END
