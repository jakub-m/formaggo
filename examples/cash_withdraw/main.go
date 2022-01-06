package main

import (
	"fmt"
	fo "formaggo/state"
)

// From https://learntla.com/introduction/example/

const (
	totalMoney = 10 + 10
	nProcesses = 2
)

func main() {
	s := State{
		AccountAlice: 10,
		AccountBob:   10,
	}
	s.P[0].Step = stepCheck
	s.P[1].Step = stepCheck
	for m0 := 1; m0 <= 20; m0++ {
		for m1 := 1; m1 <= 20; m1++ {
			s.P[0].Money = m0
			s.P[1].Money = m1

			checker := fo.Checker{
				InitialState: s,
				NamedTransitions: []fo.NamedTransition{
					{
						Name:       "ProcCheck",
						Transition: fo.Managed(ProcCheck),
					},
					{
						Name:       "ProcTransfer",
						Transition: fo.Managed(ProcTransfer),
					},
				},
				NamedInvariants: []fo.NamedInvariant{
					{
						Name: "TypeInvariant",
						Inv:  TypeInvariant,
					},
					{
						Name: "MoneyNonNegativeInvariant",
						Inv:  MoneyNonNegativeInvariant,
					},
					{
						Name: "CheckFinalBalance",
						Inv:  CheckFinalBalance,
					},
					{
						Name: "TotalMoneyInvariant",
						Inv:  TotalMoneyInvariant,
					},
				},
			}
			// TODO here reuse output graph in another iteration.
			_, violation := checker.Run()
			if violation != nil {
				for i, p := range violation.Path {
					fmt.Println(i, p)
				}
				panic(fmt.Sprint("VIOLATION ", violation.Inv.Name))
			}
		}
	}
}

type transferStep int

const (
	stepCheck transferStep = iota
	stepCanTransfer
	stepAfterTransfer
)

func (s transferStep) String() string {
	switch s {
	case stepCheck:
		return "Check"
	case stepCanTransfer:
		return "CanTransfer"
	case stepAfterTransfer:
		return "AfterTransfer"
	default:
		return "???"
	}
}

type State struct {
	AccountAlice int
	AccountBob   int
	P            [nProcesses]struct {
		Step  transferStep
		Money int
	}
}

func (s State) String() string {
	return fmt.Sprintf("{a:%d, b:%d, p0: %+v, p1: %+v}", s.AccountAlice, s.AccountBob, s.P[0], s.P[1])
}

func ProcCheck(sm *fo.StateManager) {
	curr := sm.Curr().(State)
	for i := 0; i < nProcesses; i++ {
		next := curr
		p := &next.P[i]
		if p.Step == stepCheck &&
			curr.AccountAlice >= p.Money {
			p.Step = stepCanTransfer
		}
		sm.AddNextState(next)
	}
}

func ProcTransfer(sm *fo.StateManager) {
	curr := sm.Curr().(State)
	for i := 0; i < nProcesses; i++ {
		next := curr
		p := &next.P[i]
		if p.Step == stepCanTransfer {
			next.AccountAlice -= p.Money
			next.AccountBob += p.Money
			p.Step = stepAfterTransfer
		}
		sm.AddNextState(next)
	}
}

// func ProcCheckBalance(in interface{}) []interface{} {
// 	curr := in.(State)
// 	return []interface{}{curr}
// }

func TypeInvariant(currI, nextI interface{}) bool {
	curr := currI.(State)
	return true &&
		curr.AccountAlice >= 0 &&
		curr.AccountAlice <= totalMoney &&
		curr.AccountBob >= 0 &&
		curr.AccountBob <= totalMoney
}

func MoneyNonNegativeInvariant(currI, nextI interface{}) bool {
	curr := currI.(State)
	return curr.AccountAlice >= 0 && curr.AccountBob >= 0
}

func TotalMoneyInvariant(currI, nextI interface{}) bool {
	curr := currI.(State)
	return curr.AccountAlice+curr.AccountBob == totalMoney
}

func CheckFinalBalance(currI, nextI interface{}) bool {
	curr := currI.(State)
	for i := 0; i < nProcesses; i++ {
		p := &curr.P[i]
		if p.Step != stepAfterTransfer {
			return true
		}
	}
	return curr.AccountAlice >= 0 && curr.AccountBob >= 0
}
