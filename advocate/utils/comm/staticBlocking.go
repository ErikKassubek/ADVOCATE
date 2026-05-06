// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: process.go
// Brief: Functions to communicate between runtime and advocate
//
// Author: Erik Kassubek
// Created: 2026-04-20
//
// License: BSD-3-Clause

package comm

import (
	"advocate/static/blockingStatic"
	"advocate/utils/log"
)

func (self *Communication) staticBlocking() {
	data, err := self.Get()
	if err != nil {
		log.Error(err)
	}

	res := blockingStatic.RunDynamicBlockingAnalysis(data)

	err = self.Post(res)
	if err != nil {
		log.Error(err)
	}
}
