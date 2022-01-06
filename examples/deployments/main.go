// A variation of modelling deployments from
// https://www.hillelwayne.com/post/modeling-deployments/

package main

import (
	"encoding/json"
	"fmt"
	sta "formaggo/state"
)

const (
	nServers = 2 // works for 3 as well
)

func main() {
	initialState := State{}
	iterServers(func(i int) {
		initialState.LoadBalancer[i] = true
		initialState.Version[i] = UpdateState_OldVersion
	})

	checker := sta.Checker{
		InitialState: initialState,
		NamedTransitions: []sta.NamedTransition{
			{
				Name:       "RemoveFromLoadBalancer",
				Transition: sta.Managed(RemoveFromLoadBalancer),
			},
			{
				Name:       "FlagForUpdate",
				Transition: sta.Managed(FlagForUpdate),
			},
			{
				Name:       "StartUpdate",
				Transition: sta.Managed(StartUpdate),
			},
			{
				Name:       "FinishUpdate",
				Transition: sta.Managed(FinishUpdate),
			},
			{
				Name:       "FlipLoadBalancer",
				Transition: sta.Managed(FlipLoadBalancer),
			},
			{
				Name:       "EnableLoadBalancer",
				Transition: sta.Managed(EnableLoadBalancer),
			},
		},
		NamedInvariants: []sta.NamedInvariant{
			{
				Name: "SameVersionInvariant",
				Inv:  SameVersionInvariant,
			},
			{
				Name: "ZeroDowntimeInvariant",
				Inv:  ZeroDowntimeInvariant,
			},
			{
				Name: "LoadBalancerNeverDownInvariant",
				Inv:  LoadBalancerNeverDownInvariant,
			},
		},
		NamedProperties: []sta.NamedTemporalProperty{
			{
				Name: "PropAllDeployed",
				Property: sta.TemporalProperty{
					Prop:     sta.CheckReachesAndStays,
					Initial:  sta.StateEquals(initialState),
					Terminal: PropAllDeployed,
				},
			},
		},
	}

	graph, violation := checker.Run()
	fmt.Printf("Number of states: %d\n", graph.NumStates())

	if violation != nil {
		fmt.Print(violation)
	}
	//graph.ExportToDotFile("deployment.dot")
}

// Arrays of fixed size are very practical here. Such State has no heap referenes, and does not need any "copy"
// operations. Passing the state as-is to the transformations "copies" it, since it is all value-based.
type State struct {
	UpdatePhase  int
	LoadBalancer [nServers]bool
	UpdateFlag   [nServers]bool
	Version      [nServers]UpdateState
}

func (s State) LoadBalancerFullyEnabled() bool {
	allEnabled := true
	iterServers(func(i int) {
		allEnabled = allEnabled && s.LoadBalancer[i]
	})
	return allEnabled
}

func (s State) AllUpToDate() bool {
	allUpdated := true
	iterServers(func(i int) {
		allUpdated = allUpdated && s.Version[i] == UpdateState_NewVersion
	})
	return allUpdated
}

func RemoveFromLoadBalancer(sm *sta.StateManager) {
	curr := sm.Curr().(State)
	if curr.LoadBalancerFullyEnabled() && !curr.AllUpToDate() {
		iterServers(func(i int) {
			next := curr
			next.LoadBalancer[i] = false
			sm.AddNextState(next)
		})
	}
}

func FlagForUpdate(sm *sta.StateManager) {
	curr := sm.Curr().(State)
	iterServers(func(i int) {
		next := curr
		if !curr.LoadBalancer[i] && curr.Version[i] == UpdateState_OldVersion && !curr.UpdateFlag[i] {
			next.UpdateFlag[i] = true
		}
		sm.AddNextState(next)
	})
}

func StartUpdate(sm *sta.StateManager) {
	curr := sm.Curr().(State)
	iterServers(func(i int) {
		next := curr
		if curr.UpdateFlag[i] {
			next.Version[i] = UpdateState_Updating
			next.UpdateFlag[i] = false
		}
		sm.AddNextState(next)
	})
}

func FinishUpdate(sm *sta.StateManager) {
	curr := sm.Curr().(State)
	iterServers(func(i int) {
		next := curr
		if curr.Version[i] == UpdateState_Updating {
			next.Version[i] = UpdateState_NewVersion
		}
		sm.AddNextState(next)
	})
}

func FlipLoadBalancer(sm *sta.StateManager) {
	curr := sm.Curr().(State)
	next := curr
	allDisabledAreNew := false
	someEnabledAreOld := false

	iterServers(func(i int) {
		allDisabledAreNew = allDisabledAreNew ||
			(!curr.LoadBalancer[i] && curr.Version[i] == UpdateState_NewVersion)

		someEnabledAreOld = someEnabledAreOld ||
			(curr.LoadBalancer[i] && curr.Version[i] == UpdateState_OldVersion)

	})

	if allDisabledAreNew && someEnabledAreOld {
		iterServers(func(i int) {
			next.LoadBalancer[i] = !next.LoadBalancer[i]
		})
	}
	sm.AddNextState(next)
}

func EnableLoadBalancer(sm *sta.StateManager) {
	curr := sm.Curr().(State)
	if curr.AllUpToDate() {
		next := curr
		iterServers(func(i int) {
			next.LoadBalancer[i] = true
		})
		sm.AddNextState(next)
	}
}

func SameVersionInvariant(currI, nextI interface{}) bool {
	curr := currI.(State)
	m := make(map[UpdateState]bool)
	iterServers(func(i int) {
		if curr.LoadBalancer[i] {
			m[curr.Version[i]] = true
		}
	})
	return len(m) <= 1
}

func ZeroDowntimeInvariant(currI, nextI interface{}) bool {
	curr := currI.(State)
	for i := 0; i < nServers; i++ {
		if curr.LoadBalancer[i] && curr.Version[i] == UpdateState_Updating {
			return false
		}
	}
	return true
}

func LoadBalancerNeverDownInvariant(currI, nextI interface{}) bool {
	anyUp := false
	iterServers(func(i int) {
		anyUp = anyUp || currI.(State).LoadBalancer[i]
	})
	return anyUp
}

type UpdateState string

const (
	UpdateState_OldVersion UpdateState = "old"
	UpdateState_Updating   UpdateState = "updating"
	UpdateState_NewVersion UpdateState = "new"
)

func (s UpdateState) String() string {
	return string(s)
}

func (s State) String() string {
	if j, err := json.Marshal(s); err == nil {
		return string(j)
	} else {
		return fmt.Sprint(err)
	}
}

func iterServers(fn func(i int)) {
	for i := 0; i < nServers; i++ {
		fn(i)
	}
}

func PropAllDeployed(i interface{}) bool {
	curr := i.(State)
	return curr.LoadBalancerFullyEnabled() && curr.AllUpToDate()
}
