// Copyright (c) 2024 Erik Kassubek
//
// File: analysisGraph.go
// Brief: Functions to use a graph for analysis. Used fop negative wait group
//   counter and unlock before lock
//
// Author: Erik Kassubek
// Created: 2024-09-23
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"fmt"
	"math"
)

var (
	source = &TraceElementWait{id: -1}
	drain  = &TraceElementWait{id: -2}
)

// TODO: change to graph of elems, not tID

// Build a st graph for a wait group.
// The graph has the following structure:
// - a start node s
// - a end node t
// - edges from s to all done operations
// - edges from all add operations to t
// - edges from done to add if the add happens before the done
//
// Parameter:
//   - increases map[int][]TraceElement: Operations that increase the "counter" (adds and locks)
//   - decreases map[int][]TraceElement: Operations that decrease the "counter" (dones and unlocks)
//
// Returns:
//   - []Edge: The graph
func buildResidualGraph(increases []TraceElement, decreases []TraceElement) map[TraceElement][]TraceElement {
	graph := make(map[TraceElement][]TraceElement, 0)
	graph[source] = []TraceElement{}
	graph[drain] = []TraceElement{}

	// add edges from s to all done operations
	for _, elem := range decreases {
		graph[elem] = []TraceElement{}
		graph[source] = append(graph[source], elem)
	}

	// add edges from all add operations to t
	for _, elem := range increases {
		graph[elem] = []TraceElement{drain}

	}

	// add edge from done to add if the add happens before the done
	for _, elemDecrease := range decreases {
		for _, elemIncrease := range increases {
			if clock.GetHappensBefore(elemIncrease.GetVC(), elemDecrease.GetVC()) == clock.Before {
				graph[elemDecrease] = append(graph[elemDecrease], elemIncrease)
			}
		}
	}

	return graph
}

// Calculate the maximum flow of a graph using the ford fulkerson algorithm
//
// Parameter:
//   - graph map[TraceElement][]TraceElement: The graph
//
// Returns:
//   - int: The maximum flow
//   - map[TraceElement][]TraceElement: The graph with max flow
//   - error
func calculateMaxFlow(graph map[TraceElement][]TraceElement) (int, map[TraceElement][]TraceElement, error) {
	maxFlow := 0
	maxNumberRounds := 0
	for _, val := range graph {
		maxNumberRounds += len(val)
	}
	maxNumberRounds = 1e5 * int(math.Pow(float64(maxNumberRounds), 3.))

	for i := 0; i < int(maxNumberRounds); i++ { // max number rounds to prevent infinite loop
		path, flow := findPath(graph)
		if flow == 0 {
			return maxFlow, graph, nil
		}

		maxFlow += flow
		for i := 0; i < len(path)-1; i++ {
			graph[path[i]] = append(graph[path[i]], path[i+1])
			graph[path[i+1]] = remove(graph[path[i+1]], path[i])
		}
	}

	return maxFlow, graph, fmt.Errorf("To many rounds")
}

// Find a path in a graph using a breadth-fifoirst search
//
// Parameter:
//   - map[TraceElement][]TraceElement
//
// Returns:
//   - []TraceElement: The path
//   - int: The flow
func findPath(graph map[TraceElement][]TraceElement) ([]TraceElement, int) {
	visited := make(map[TraceElement]bool, 0)

	queue := []TraceElement{source}
	visited[source] = true
	parents := make(map[TraceElement]TraceElement, 0)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node.IsEqual(drain) {
			path := []TraceElement{}
			for !node.IsEqual(source) {
				path = append(path, node)
				node = parents[node]
			}
			path = append(path, source)

			return path, 1
		}

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
				parents[neighbor] = node
			}
		}
	}

	return []TraceElement{}, 0
}

// Remove an element from a list
//
// Parameter:
//   - list []TraceElement: The list
//   - element TraceElement: The element to remove
//
// Returns:
//   - []string: The list without the element
func remove(list []TraceElement, element TraceElement) []TraceElement {
	for i, e := range list {
		if element.IsEqual(e) {
			list = append(list[:i], list[i+1:]...)
			return list
		}
	}
	return list
}
