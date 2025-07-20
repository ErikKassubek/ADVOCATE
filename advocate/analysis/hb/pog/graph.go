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
	return poGraph{map[trace.Element]map[trace.Element]struct{}{}}
}

func InitPOG() {
	bufferedVCs = make(map[int]([]data.BufferedVC))
	bufferedVCsCount = make(map[int]int)
	bufferedVCsSize = make(map[int]int)

	po = newPoGraph()
	poInverted = newPoGraph()

	poWeak = newPoGraph()
	poWeakInverted = newPoGraph()
}

func (g *poGraph) addEdge(from, to trace.Element) {
	if _, ok := g.data[from]; !ok {
		g.data[from] = make(map[trace.Element]struct{})
	}
	g.data[from][to] = struct{}{}
}

func (g *poGraph) getChildren(from trace.Element) map[trace.Element]struct{} {
	return g.data[from]
}

func Print(weak bool) {
	if weak {
		log.Important(poWeak.toString())
	} else {
		log.Important(po.toString())
	}
}

func (g *poGraph) toString() string {
	res := ""
	for start, end := range g.data {
		res += fmt.Sprintf("%d -> ", start.GetLine())
		for e, _ := range end {
			res += fmt.Sprintf("%d, ", e.GetLine())
		}
	}
	return res
}
