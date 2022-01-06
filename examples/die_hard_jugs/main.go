package main

import (
	"fmt"
	fo "formaggo/state"
)

// You have a 3-gallon and a 5-gallon jug that you can fill from a fountain of water.  Fill one of the jugs with
// exactly 4 gallons of water.

func main() {

	checker := fo.Checker{
		InitialState: Jugs{Jug3: 0, Jug5: 0},
		NamedTransitions: []fo.NamedTransition{
			{
				Transition: fo.Managed(EmptyOrFillJugs),
				Name:       "EmptyOrFillJugs",
			},
			{
				Transition: fo.Managed(PourJug3ToJug5),
				Name:       "PourJug3ToJug5",
			},
			{
				Transition: fo.Managed(PourJug5ToJug3),
				Name:       "PourJug5ToJug3",
			},
		},
		NamedInvariants: []fo.NamedInvariant{
			{
				Name: "InvJugSize",
				Inv:  InvJugSize,
			},
			{
				Name: "InvEndCondition",
				Inv:  InvEndCondition,
			},
			{
				Name: "InvConstantWater",
				Inv:  InvConstantWater,
			},
		},
	}

	_, violation := checker.Run()
	if violation != nil {
		fmt.Println(violation.Inv.Name)
		for i, p := range violation.Path {
			fmt.Println(i, p)
		}
	}
}

type Jugs struct {
	Jug3 int
	Jug5 int
}

func (s Jugs) String() string {
	return fmt.Sprintf("{j3: %d, j5: %d}", s.Jug3, s.Jug5)
}

func EmptyOrFillJugs(sm *fo.StateManager) {
	curr := sm.Curr().(Jugs)
	if curr.Jug3 > 0 {
		next := curr
		next.Jug3 = 0
		sm.AddNextState(next)
	}
	if curr.Jug5 > 0 {
		next := curr
		next.Jug5 = 0
		sm.AddNextState(next)
	}
	if curr.Jug3 < 3 {
		next := curr
		next.Jug3 = 3
		sm.AddNextState(next)
	}
	if curr.Jug5 < 5 {
		next := curr
		next.Jug5 = 5
		sm.AddNextState(next)
	}
}

func PourJug5ToJug3(sm *fo.StateManager) {
	curr := sm.Curr().(Jugs)
	if curr.Jug5 > 0 && curr.Jug3 < 3 {
		next := curr
		space := 3 - curr.Jug3
		if space >= curr.Jug5 {
			// all will fit
			next.Jug5 = 0
			next.Jug3 += curr.Jug5
			sm.AddNextState(next)
		} else {
			next.Jug5 -= space
			next.Jug3 = 3
			sm.AddNextState(next)
		}
	}
}

func PourJug3ToJug5(m *fo.StateManager) {
	curr := m.Curr().(Jugs)
	if curr.Jug3 > 0 && curr.Jug5 < 5 {
		next := curr
		space := 5 - curr.Jug5
		if curr.Jug3 < space {
			next.Jug3 = 0
			next.Jug5 += curr.Jug3
			m.AddNextState(next)
		} else {
			next.Jug3 -= space
			next.Jug5 = 5
			m.AddNextState(next)
		}
	}
}

func InvJugSize(currIn, nextIn interface{}) bool {
	curr := currIn.(Jugs)
	return (curr.Jug3 >= 0 && curr.Jug3 <= 3) && (curr.Jug5 >= 0 && curr.Jug5 <= 5)
}

func InvEndCondition(currIn, nextIn interface{}) bool {
	curr := currIn.(Jugs)
	return curr.Jug5 != 4
}

func InvConstantWater(currIn, nextIn interface{}) bool {
	curr, next := currIn.(Jugs), nextIn.(Jugs)
	if curr.Jug3 != next.Jug3 && curr.Jug5 != next.Jug5 {
		return curr.Jug3+curr.Jug5 == next.Jug3+next.Jug5
	}
	return true
}
