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
	"strings"
)

func (self *Communication) staticBlocking() {
	go func() {
		msg, err := self.Get()
		if err != nil {
			log.Error(err)
			return
		}

		data := strings.SplitN(msg, "?", 1)

		res := ""

		if len(data) == 2 {
			switch data[0] {
			case "STATICRELEASABLE":
				res = blockingStatic.RunDynamicBlockingAnalysis(data[1]) // TODO: return "0" if cannot releas, "1" otherwise
			default:
				res = "UNKNOWN KEY: " + data[0]
			}
		} else {
			res = "UNKNOWN MESSAGE: " + msg
		}

		err = self.Post(res)
		if err != nil {
			log.Error("Post error: ", err)
		}
	}()
}
