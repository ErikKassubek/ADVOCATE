// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

// GOCDR-START

#include "textflag.h"

TEXT ·SwapInt32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchg(SB)

TEXT ·SwapUint32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchg(SB)

TEXT ·SwapInt64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchg64(SB)

TEXT ·SwapUint64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchg64(SB)

TEXT ·SwapUintptrGocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchguintptr(SB)

TEXT ·CompareAndSwapInt32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Cas(SB)

TEXT ·CompareAndSwapUint32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Cas(SB)

TEXT ·CompareAndSwapUintptrGocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Casuintptr(SB)

TEXT ·CompareAndSwapInt64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Cas64(SB)

TEXT ·CompareAndSwapUint64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Cas64(SB)

TEXT ·AddInt32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadd(SB)

TEXT ·AddUint32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadd(SB)

TEXT ·AddUintptrGocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadduintptr(SB)

TEXT ·AddInt64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadd64(SB)

TEXT ·AddUint64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadd64(SB)

TEXT ·LoadInt32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Load(SB)

TEXT ·LoadUint32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Load(SB)

TEXT ·LoadInt64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Load64(SB)

TEXT ·LoadUint64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Load64(SB)

TEXT ·LoadUintptrGocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Loaduintptr(SB)

TEXT ·LoadPointerGocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Loadp(SB)

TEXT ·StoreInt32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Store(SB)

TEXT ·StoreUint32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Store(SB)

TEXT ·StoreInt64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Store64(SB)

TEXT ·StoreUint64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Store64(SB)

TEXT ·StoreUintptrGocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Storeuintptr(SB)

TEXT ·AndInt32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·And32(SB)

TEXT ·AndUint32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·And32(SB)

TEXT ·AndUintptrGocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Anduintptr(SB)

TEXT ·AndInt64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·And64(SB)

TEXT ·AndUint64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·And64(SB)

TEXT ·OrInt32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Or32(SB)

TEXT ·OrUint32Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Or32(SB)

TEXT ·OrUintptrGocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Oruintptr(SB)

TEXT ·OrInt64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Or64(SB)

TEXT ·OrUint64Gocdr(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Or64(SB)

// GOCDR-END