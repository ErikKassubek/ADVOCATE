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

/*
 * Build a st graph for a wait group.
 * The graph has the following structure:
 * - a start node s
 * - a end node t
 * - edges from s to all done operations
 * - edges from all add operations to t
 * - edges from done to add if the add happens before the done
 * Args:
 * 	increases (map[int][]TraceElement): Operations that increase the "counter" (adds and locks)
 * 	decreases (map[int][]TraceElement): Operations that decrease the "counter" (dones and unlocks)
 * Returns:
 * 	[]Edge: The graph
 */
func buildResidualGraph(increases []TraceElement, decreases []TraceElement) map[string][]string {
	graph := make(map[string][]string, 0)
	graph["s"] = []string{}
	graph["t"] = []string{}

	// add edges from s to all done operations
	for _, elem := range decreases {
		graph[elem.GetTID()] = []string{}
		graph["s"] = append(graph["s"], elem.GetTID())
	}

	// add edges from all add operations to t
	for _, elem := range increases {
		graph[elem.GetTID()] = []string{"t"}

	}

	// add edge from done to add if the add happens before the done
	for _, elemDecrease := range decreases {
		for _, elemIncrease := range increases {
			if clock.GetHappensBefore(elemIncrease.GetVC(), elemDecrease.GetVC()) == clock.Before {
				graph[elemDecrease.GetTID()] = append(graph[elemDecrease.GetTID()], elemIncrease.GetTID())
			}
		}
	}

	return graph
}

/*
 * Calculate the maximum flow of a graph using the ford fulkerson algorithm
 * Args:
 * 	graph ([]Edge): The graph
 * Returns:
 * 	int: The maximum flow
 */
func calculateMaxFlow(graph map[string][]string) (int, map[string][]string, error) {
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

/*
 * Find a path in a graph using a breadth-fifoirst search
 * Args:
 * 	graph ([]Edge): The graph
 * Returns:
 * 	[]string: The path
 * 	int: The flow
 */
func findPath(graph map[string][]string) ([]string, int) {
	visited := make(map[string]bool, 0)
	queue := []string{"s"}
	visited["s"] = true
	parents := make(map[string]string, 0)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node == "t" {
			path := []string{}
			for node != "s" {
				path = append(path, node)
				node = parents[node]
			}
			path = append(path, "s")

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

	return []string{}, 0
}

/*
 * Remove an element from a list
 * Args:
 * 	list ([]string): The list
 * 	element (string): The element to remove
 * Returns:
 * 	[]string: The list without the element
 */
func remove(list []string, element string) []string {
	for i, e := range list {
		if e == element {
			list = append(list[:i], list[i+1:]...)
			return list
		}
	}
	return list
}
