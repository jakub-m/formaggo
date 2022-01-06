package collection

type IntSet map[int]bool

func NewIntSet(vals ...int) *IntSet {
	s := make(IntSet)
	// TODO add unit test for that
	for _, v := range vals {
		s.Add(v)
	}
	return &s
}

func (s IntSet) Contains(val int) bool {
	if v, ok := s[val]; ok {
		return v
	} else {
		return false
	}
}

// func (s IntSet) Hash() hash.Hash {
// }

func (s IntSet) Values() []int {
	vals := []int{}
	for k, v := range s {
		if v {
			vals = append(vals, k)
		}
	}
	return vals
}

func (s IntSet) Copy() *IntSet {
	copied := NewIntSet()
	for k, v := range s {
		if v {
			(*copied)[k] = v
		}
	}
	return copied
}

func (s *IntSet) Add(val int) {
	(*s)[val] = true
}
