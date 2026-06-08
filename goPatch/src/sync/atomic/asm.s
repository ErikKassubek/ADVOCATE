// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !race

// OOSC-START

#include "textflag.h"

TEXT ·SwapInt32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchg(SB)

TEXT ·SwapUint32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchg(SB)

TEXT ·SwapInt64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchg64(SB)

TEXT ·SwapUint64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchg64(SB)

TEXT ·SwapUintptrOosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xchguintptr(SB)

TEXT ·CompareAndSwapInt32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Cas(SB)

TEXT ·CompareAndSwapUint32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Cas(SB)

TEXT ·CompareAndSwapUintptrOosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Casuintptr(SB)

TEXT ·CompareAndSwapInt64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Cas64(SB)

TEXT ·CompareAndSwapUint64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Cas64(SB)

TEXT ·AddInt32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadd(SB)

TEXT ·AddUint32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadd(SB)

TEXT ·AddUintptrOosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadduintptr(SB)

TEXT ·AddInt64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadd64(SB)

TEXT ·AddUint64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Xadd64(SB)

TEXT ·LoadInt32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Load(SB)

TEXT ·LoadUint32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Load(SB)

TEXT ·LoadInt64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Load64(SB)

TEXT ·LoadUint64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Load64(SB)

TEXT ·LoadUintptrOosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Loaduintptr(SB)

TEXT ·LoadPointerOosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Loadp(SB)

TEXT ·StoreInt32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Store(SB)

TEXT ·StoreUint32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Store(SB)

TEXT ·StoreInt64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Store64(SB)

TEXT ·StoreUint64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Store64(SB)

TEXT ·StoreUintptrOosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Storeuintptr(SB)

TEXT ·AndInt32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·And32(SB)

TEXT ·AndUint32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·And32(SB)

TEXT ·AndUintptrOosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Anduintptr(SB)

TEXT ·AndInt64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·And64(SB)

TEXT ·AndUint64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·And64(SB)

TEXT ·OrInt32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Or32(SB)

TEXT ·OrUint32Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Or32(SB)

TEXT ·OrUintptrOosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Oruintptr(SB)

TEXT ·OrInt64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Or64(SB)

TEXT ·OrUint64Oosc(SB),NOSPLIT,$0
	JMP	internal∕runtime∕atomic·Or64(SB)

// OOSC-END