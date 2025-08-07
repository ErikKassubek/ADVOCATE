//
// File: graph.go
// Brief: Store the partial order graph
//
// Created: 2025-07-08
//
// License: BSD-3-Clause

package pog

import (
	"fmt"
	"goCR/analysis/data"
	"goCR/trace"
	"goCR/utils/log"
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

func (g *poGraph) addEdge(from, to trace.Element) {
	if _, ok := g.data[from]; !ok {
		g.data[from] = make(map[trace.Element]struct{})
	}
	g.data[from][to] = struct{}{}
}

func (g *poGraph) getChildren(from trace.Element) map[trace.Element]struct{} {
	return g.data[from]
}

// Print prints the current graph
//
// Parameter:
//   - weak bool: if true, print the weak hb tree, otherwise print the string
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
		for e := range end {
			res += fmt.Sprintf("%d, ", e.GetLine())
		}
	}
	return res
}
