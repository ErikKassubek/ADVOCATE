// Copyright (c) 2025 Erik Kassubek
//
// File: graph.go
// Brief: Store the partial order graph
//
// Author: Erik Kassubek
// Created: 2025-07-08
//
// License: BSD-3-Clause

package pog

import (
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/log"
	"fmt"
)

var (
	po         poGraph
	poInverted poGraph

	poWeak         poGraph
	poWeakInverted poGraph
)

type poGraph struct {
	data map[trace.Element]map[trace.Element]struct{}
}

func newPoGraph() poGraph {
	return poGraph{make(map[trace.Element]map[trace.Element]struct{})}
}

// InitPOG initializes the directed acyclic partial order graph
func InitPOG() {
	chanBuffer = make(map[int]([]data.BufferedVC))
	chanBufferSize = make(map[int]int)

	po = newPoGraph()
	poInverted = newPoGraph()

	poWeak = newPoGraph()
	poWeakInverted = newPoGraph()
}

func (this *poGraph) addEdge(from, to trace.Element) {
	if _, ok := this.data[from]; !ok {
		this.data[from] = make(map[trace.Element]struct{})
	}
	this.data[from][to] = struct{}{}
}

func (this *poGraph) getChildren(from trace.Element) map[trace.Element]struct{} {
	return this.data[from]
}

// Print prints the current graph
//
// Parameter:
//   - weak bool: if true, print the weak hb tree, otherwise print the string
func Print(weak bool) {
	if weak {
		log.Info(poWeak.toString())
	} else {
		log.Info(po.toString())
	}
}

func (this *poGraph) toString() string {
	res := ""
	for start, end := range this.data {
		res += fmt.Sprintf("%d -> ", start.GetLine())
		for e := range end {
			res += fmt.Sprintf("%d, ", e.GetLine())
		}
	}
	return res
}
