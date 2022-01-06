package state

// StateManager is a convenience structure taht can (but does not have to) wrap a Transition. With StateManager
// one can populate next states with AddNextState.
type StateManager struct {
	// TODO check that curr holds same hash, not mutated.
	curr interface{}
	next []interface{}
}

func (s *StateManager) Curr() interface{} {
	return s.curr
}

func (s *StateManager) AddNextState(next interface{}) {
	s.next = append(s.next, next)
}

// Managed wraps Transition, so the implementation of the Transition uses StateManager.
func Managed(tran func(*StateManager)) Transition {
	return func(curr interface{}) []interface{} {
		sm := StateManager{
			curr: curr,
			next: []interface{}{curr},
		}
		tran(&sm)
		return sm.next
	}
}
