package state

import (
	"fmt"

	"github.com/jakub-m/formaggo/log"

	spa "github.com/jakub-m/formaggo/shortestpath"

	"github.com/mitchellh/hashstructure/v2"
)

type Checker struct {
	// InitialState is the starting state of the analysis.
	InitialState interface{}
	// Transitions are next-state-relation functions. There must be at least one transition.
	NamedTransitions []NamedTransition
	// Invariants must hold for each analysed state. Optional.
	NamedInvariants []NamedInvariant
	// NamedProperties must hold for all the possible paths in the state transition graph. Optional
	NamedProperties []NamedTemporalProperty
}

type StateGraph struct {
	hashGraph   map[stateHash][]stateHash
	hashToState map[stateHash]interface{}
}

func (g StateGraph) NumStates() int {
	return len(g.hashToState)
}

type NamedTemporalProperty struct {
	Name     string
	Property TemporalProperty
}

type TemporalProperty struct {
	Prop     func(g StateGraph, initial, terminal StateCondition) []interface{}
	Initial  StateCondition
	Terminal StateCondition
}

type NamedTransition struct {
	Name       string
	Transition Transition
}

func (c Checker) Run() (StateGraph, *Violation) {
	log.Println("Start checker")
	graph, violation := c.runTransitions()
	if violation != nil {
		return graph, violation
	}
	log.Printf("Done generating graph of size: %d\n", graph.NumStates())
	violation = c.runTemporalChecks(graph)
	return graph, violation
}

func (c Checker) runTransitions() (StateGraph, *Violation) {
	sg := StateGraph{
		hashGraph:   make(map[stateHash][]stateHash),
		hashToState: make(map[stateHash]interface{}),
	}

	initialState := c.InitialState
	log.Debugln("init", initialState)
	initialStateHash := GetHash(initialState)
	sg.hashToState[initialStateHash] = initialState
	backlog := []stateHash{initialStateHash}
	for len(backlog) > 0 {
		log.Debugf("backlog %d", len(backlog))
		currHash := backlog[len(backlog)-1]
		backlog = backlog[0 : len(backlog)-1]

		if _, ok := sg.hashGraph[currHash]; ok {
			// The state was already processed, all the transisitons are in the map. No need to do it again.
			log.Debugf("continue")
			continue
		}

		curr, ok := sg.hashToState[currHash]
		if !ok {
			panic(fmt.Sprintf("RATS! hashToState does not have a corresponding entry: %v", currHash))
		}
		log.Debugf("curr %v", curr)

		nextStateHashSet := make(map[stateHash]interface{})
		for _, namTran := range c.NamedTransitions {
			statesAfterTransition := namTran.Transition(curr)
			if len(statesAfterTransition) == 0 {
				panic(fmt.Sprintf("there are no future states after transition %+v", namTran.Name)) // Make it an error. Named transitions? Reflection?
			}
			for _, next := range statesAfterTransition {
				nextHash := GetHash(next)
				nextStateHashSet[nextHash] = next
			}
		}

		nextStateHashes := []stateHash{}
		for nextHash, next := range nextStateHashSet {
			sg.hashToState[nextHash] = next
			log.Debugf("%v -> %v", curr, next)

			nextStateHashes = append(nextStateHashes, nextHash)

			for _, namInv := range c.NamedInvariants {
				if !namInv.Inv(curr, next) {
					path := []interface{}{}
					for _, h := range findShortestPathHash(sg.hashGraph, initialStateHash, currHash) {
						path = append(path, sg.hashToState[h])
					}
					violation := Violation{
						Inv:              &namInv,
						Curr:             curr,
						Next:             next,
						Path:             path,
						namedTransitions: c.NamedTransitions,
					}
					return sg, &violation
				}
			}

			if _, ok := sg.hashGraph[nextHash]; !ok {
				// Minor optimization. Do not fill backlog with the states that will be skipped immediately at the beginning of the loop.
				backlog = append(backlog, nextHash)
			}
		}
		sg.hashGraph[currHash] = nextStateHashes
	}

	log.Debugf("all hashshes: %d", len(sg.hashToState))
	return sg, nil
}

func (c Checker) runTemporalChecks(g StateGraph) *Violation {
	log.Println("Now check temporal properties")
	for _, prop := range c.NamedProperties {
		log.Printf("%s\n", prop.Name)
		if counterExample := prop.Property.Prop(g, prop.Property.Initial, prop.Property.Terminal); counterExample != nil {
			return &Violation{
				Prop:             &prop,
				Path:             counterExample,
				namedTransitions: c.NamedTransitions,
			}
		}
	}
	return nil
}

func findShortestPathHash(stateHashTransitionMap map[stateHash][]stateHash, start, end stateHash) []stateHash {
	verStart := spa.Vertex(start)
	verEnd := spa.Vertex(end)

	getNext := func(currVertex spa.Vertex) []spa.Vertex {
		nextVertices := []spa.Vertex{}
		for _, nextStateHash := range stateHashTransitionMap[stateHash(currVertex)] {
			nextVertices = append(nextVertices, spa.Vertex(nextStateHash))
		}
		return nextVertices
	}

	hashPath := []stateHash{}

	for _, h := range spa.Find(verStart, verEnd, getNext) {
		hashPath = append(hashPath, stateHash(h))
	}

	return hashPath
}

type Analysis struct {
	Violation              *Violation
	stateHashTransitionMap map[stateHash][]stateHash
	hashToStateMap         map[stateHash]interface{}
}

type Violation struct {
	Inv              *NamedInvariant
	Prop             *NamedTemporalProperty
	Curr, Next       interface{}
	Path             []interface{}
	namedTransitions []NamedTransition
}

func (v Violation) String() string {
	s := ""
	if v.Inv != nil {
		s += fmt.Sprintf("Violation of invariant: %s\n", v.Inv.Name)
	}
	if v.Prop != nil {
		s += fmt.Sprintf("Violation of property: %s\n", v.Prop.Name)
	}
	if v.Curr != nil {
		s += fmt.Sprintf("Current: %s\n", v.Curr)
	}
	if v.Next != nil {
		s += fmt.Sprintf("Next:    %s\n", v.Curr)
	}
	if v.Path != nil && len(v.Path) > 0 {
		lastHash := GetHash(v.Path[len(v.Path)-1])
		loopIndex := -1
		for i, stateOnPath := range v.Path {
			s += fmt.Sprintf("%d\t%s\n", i, stateOnPath)
			if GetHash(stateOnPath) == lastHash && i != len(v.Path)-1 {
				loopIndex = i
			}
			if i != len(v.Path)-1 {
				if exp, ok := v.findTransitionMatchingStates(v.Path[i], v.Path[i+1]); ok {
					s += fmt.Sprintf("%d->%d\t%s\n", i, i+1, exp)
				} else {
					s += fmt.Sprintf("%d->%d\tRATS! Illegal transition\n", i, i+1)
				}
			}
		}
		if loopIndex != -1 {
			s += fmt.Sprintf("(back to %d)\n", loopIndex)
		}
	}
	return s
}

func (v Violation) findTransitionMatchingStates(curr, next interface{}) (string, bool) {
	for _, t := range v.namedTransitions {
		for _, tentativeNext := range t.Transition(curr) {
			if GetHash(next) == GetHash(tentativeNext) {
				return t.Name, true
			}
		}
	}
	return "", false
}

type Transition func(interface{}) []interface{}

type NamedInvariant struct {
	Name string
	Inv  Invariant
}

type Invariant func(curr, next interface{}) bool

type stateHash uint64

func GetHash(in interface{}) stateHash {
	h, err := hashstructure.Hash(in, hashstructure.FormatV2, nil)
	if err != nil {
		panic(fmt.Sprint("hashOf", err))
	}
	return stateHash(h)
}
