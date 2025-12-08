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
	"advocate/analysis/baseA"
	"advocate/trace"
	"advocate/utils/log"
	"fmt"
)

var (
	po         PoGraph
	poInverted PoGraph

	poWeak         PoGraph
	poWeakInverted PoGraph
)

type PoGraph struct {
	data map[trace.Element]map[trace.Element]struct{}
}

func NewPoGraph() PoGraph {
	return PoGraph{make(map[trace.Element]map[trace.Element]struct{})}
}

// InitPOG initializes the directed acyclic partial order graph
func InitPOG() {
	chanBuffer = make(map[int]([]baseA.BufferedVC))
	chanBufferSize = make(map[int]int)

	po = NewPoGraph()
	poInverted = NewPoGraph()

	poWeak = NewPoGraph()
	poWeakInverted = NewPoGraph()
}

func (this *PoGraph) AddEdge(from, to trace.Element) {
	if _, ok := this.data[from]; !ok {
		this.data[from] = make(map[trace.Element]struct{})
	}
	this.data[from][to] = struct{}{}
}

func (this *PoGraph) GetChildren(from trace.Element) map[trace.Element]struct{} {
	return this.data[from]
}

func (this *PoGraph) IsEmpty() bool {
	return len(this.data) == 0
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

func (this *PoGraph) toString() string {
	res := ""
	for start, end := range this.data {
		res += fmt.Sprintf("%d -> ", start.GetLine())
		for e := range end {
			res += fmt.Sprintf("%d, ", e.GetLine())
		}
	}
	return res
}
