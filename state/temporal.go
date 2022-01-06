// TODO:
// - Temporal property should hold regardless of the input node. Evaluate
//   it for all the nodes and cache heavily.

package state

type StateCondition func(interface{}) bool

func Not(sc StateCondition) StateCondition {
	return func(i interface{}) bool {
		return !sc(i)
	}
}

// StateEquals works only for values, not for references.
func StateEquals(expected interface{}) StateCondition {
	return func(i interface{}) bool {
		return i == expected
	}
}

type stateHashCondition func(stateHash) bool

type TemporalPropertyFunc func(initial, terminal StateCondition) []interface{}

// CheckReachesAndStays check if every path starting at a node meeting the initial condition, always reaches node
// meeting terminal condition and stays in that state. Return nil or example of a path that violates the condidion.
func CheckReachesAndStays(g StateGraph, initial, terminal StateCondition) []interface{} {
	path := checkHashAlwaysReachesAndStaysCond(
		g.hashGraph,
		func(sh stateHash) bool {
			return initial(g.hashToState[sh])
		},
		func(sh stateHash) bool {
			return terminal(g.hashToState[sh])
		},
	)
	if path == nil {
		return nil
	}
	statePath := []interface{}{}
	for _, sh := range path {
		statePath = append(statePath, g.hashToState[sh])
	}
	return statePath
}

// checkAlwaysReachesAndStays checks that all the paths starting at the nodes meeting the initial condition always reach
// the terminal nodes, and stay there. I.e. they don't flip between non-terminal and terminal.
func checkHashAlwaysReachesAndStaysCond(transMap map[stateHash][]stateHash, initial, terminal stateHashCondition) []stateHash {
	var counterExample []stateHash
	for h := range transMap {
		if initial(h) {
			// To optimize: cache results if all possible paths starting from node X meet a condition.
			onEveryFinishedPath(transMap, h, func(path []stateHash, cycle int) bool {
				if !pathReachesAndStays(path, cycle, terminal) {
					counterExample = path
				}
				// stop on first counter example
				return counterExample == nil
			})
		}
	}
	return counterExample
}

// onEveryFinishedPath runs fn on every possible finished path. A finished path is a path
// that cannot expand any further or is a cycle. The path with a cycle will end with a
// node that is already present once in the preceding nodes. fn returns a "should continue"
// flag.
func onEveryFinishedPath(transMap map[stateHash][]stateHash, start stateHash, fn func([]stateHash, int) bool) {
	var rec func([]stateHash, int) bool
	// If cycle is other than -1 then it indicates that the last element of the path is repeated
	// at position path[cycle].
	rec = func(path []stateHash, cycle int) bool {
		last := path[len(path)-1]
		//neighbours := transMap[last] // with stuttering
		neighbours := []stateHash{}
		for _, s := range transMap[last] {
			if s != last { // no stutterning
				neighbours = append(neighbours, s)
			}
		}

		if len(neighbours) == 0 || cycle != -1 {
			// path finished - cannot expand, or is a cycle
			return fn(path, cycle)
		}

		for _, neigh := range neighbours {
			// To optimize: modify path in-place intead of creating it anew.
			newPath := append(path, neigh)
			shouldContinue := rec(newPath, findState(neigh, path))
			if !shouldContinue {
				return shouldContinue
			}
		}
		return true
	}
	rec([]stateHash{start}, -1)
}

func findState(s stateHash, path []stateHash) int {
	// To optimize: do not sweep the path in O(n) but use an O(1) set.
	for i, v := range path {
		if v == s {
			return i
		}
	}
	return -1
}

func pathReachesAndStays(path []stateHash, cycle int, terminal stateHashCondition) bool {
	if cycle == -1 {
		// no cycle, check only the last state
		return terminal(path[len(path)-1])
	} else {
		// the whole cycle loop should meet the terminal condition
		for i := cycle; i < len(path); i++ {
			if !terminal(path[i]) {
				return false
			}
		}
		return true
	}
}
