package shortestpath

import (
	"container/heap"
	"fmt"
)

type Vertex uint64

func Find(start, end Vertex, getNext func(Vertex) []Vertex) []Vertex {
	backlog := &backlogHeap{{start, 0}}
	distances := make(map[Vertex]int)
	visited := make(map[Vertex]bool)
	backHops := make(map[Vertex][]Vertex) // this is to trace back the path

	distances[start] = 0
	visited[start] = true

	// Dijkstra
	for len(*backlog) > 0 {
		current := heap.Pop(backlog).(vertexWithCost).vertex
		var currentDistance int
		if dist, ok := distances[current]; ok {
			currentDistance = dist
		} else {
			panic(fmt.Sprintf("RATS! No distance in map for vertex %v", current))
		}

		for _, next := range getNext(Vertex(current)) {
			if visited[next] {
				continue
			}
			tentativeDist := currentDistance + 1 // graph not weighted
			if dist, ok := distances[next]; ok {
				if tentativeDist < dist {
					distances[next] = tentativeDist
				}
			} else {
				distances[next] = tentativeDist
			}

			heap.Push(backlog, vertexWithCost{vertex: next, cost: distances[next]})
			if _, ok := backHops[next]; !ok {
				backHops[next] = []Vertex{}
			}
			backHops[next] = append(backHops[next], current)
		}
		visited[current] = true
	}

	// Trace back the path.
	if distances[end] == 0 {
		// not found
		return []Vertex{}
	}

	path := []Vertex{end}
	current := end
	for current != start {
		back := backHops[current]
		if len(back) == 0 {
			panic(fmt.Sprintf("RATS! No back-hops for vertex: %v", back))
		}
		selectedPrev := back[0]
		for _, candidatePrev := range back {
			if distances[candidatePrev] < distances[selectedPrev] {
				selectedPrev = candidatePrev
			}
		}

		path = append(path, selectedPrev)
		current = selectedPrev
	}

	// reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

type vertexWithCost struct {
	vertex Vertex
	cost   int
}

type backlogHeap []vertexWithCost

func (h backlogHeap) Len() int {
	return len(h)
}

func (h backlogHeap) Less(i, j int) bool {
	return h[i].cost < h[j].cost
}

func (h backlogHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *backlogHeap) Push(x interface{}) {
	*h = append(*h, x.(vertexWithCost))
}

func (h *backlogHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
