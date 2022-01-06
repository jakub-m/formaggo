package state

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIteratePaths1(t *testing.T) {
	tm := make(map[stateHash][]stateHash)
	tm[0] = []stateHash{1, 2}
	tm[1] = []stateHash{0}
	tm[2] = []stateHash{0}
	col := newCollector()
	onEveryFinishedPath(tm, 0, col.fn)
	assert.Equal(t,
		[]string{
			"0,1,0:0",
			"0,2,0:0"},
		col.sortPaths())
}

func TestIteratePaths2(t *testing.T) {
	tm := make(map[stateHash][]stateHash)
	tm[0] = []stateHash{1, 2}
	tm[1] = []stateHash{3}
	tm[2] = []stateHash{3}
	col := newCollector()
	onEveryFinishedPath(tm, 0, col.fn)
	assert.Equal(t,
		[]string{"0,1,3", "0,2,3"},
		col.sortPaths())
}

func TestIteratePaths3(t *testing.T) {
	tm := make(map[stateHash][]stateHash)
	tm[0] = []stateHash{0, 1}
	tm[1] = []stateHash{0, 1, 2}
	tm[2] = []stateHash{1, 2, 3}
	tm[3] = []stateHash{2, 3}
	col := newCollector()
	onEveryFinishedPath(tm, 0, col.fn)
	assert.Equal(t,
		[]string{
			// "0,0:0",
			"0,1,0:0",
			// "0,1,1:1",
			"0,1,2,1:1",
			// "0,1,2,2:2",
			"0,1,2,3,2:2",
			// "0,1,2,3,3:3",
		},
		col.sortPaths())
}

func TestIteratePaths4(t *testing.T) {
	tm := make(map[stateHash][]stateHash)
	tm[0] = []stateHash{1, 2, 3}
	tm[1] = []stateHash{3}
	tm[2] = []stateHash{1, 3}
	tm[3] = []stateHash{}
	col := newCollector()
	onEveryFinishedPath(tm, 0, col.fn)
	assert.Equal(t,
		[]string{
			"0,1,3",
			"0,2,1,3",
			"0,2,3",
			"0,3",
		},
		col.sortPaths())
}

func TestPathStays1(t *testing.T) {
	tm := make(map[stateHash][]stateHash)
	tm[0] = []stateHash{1}
	tm[1] = []stateHash{2}
	tm[2] = []stateHash{10}
	p := checkHashAlwaysReachesAndStaysCond(tm, eq0, gt10)
	assert.Nil(t, p, "bad: "+pathToString(p))
}

func TestPathStays2(t *testing.T) {
	tm := make(map[stateHash][]stateHash)
	tm[0] = []stateHash{1}
	tm[1] = []stateHash{2}
	tm[2] = []stateHash{10}
	tm[2] = []stateHash{3}
	p := checkHashAlwaysReachesAndStaysCond(tm, eq0, gt10)
	assert.Equal(t, []stateHash{0, 1, 2, 3}, p, "bad: "+pathToString(p))
}

func TestPathStays3(t *testing.T) {
	tm := make(map[stateHash][]stateHash)
	tm[0] = []stateHash{1}
	tm[1] = []stateHash{2}
	tm[2] = []stateHash{10}
	tm[10] = []stateHash{1}
	p := checkHashAlwaysReachesAndStaysCond(tm, eq0, gt10)
	assert.Equal(t, []stateHash{0, 1, 2, 10, 1}, p, "bad: "+pathToString(p))
}

func TestPathStays4(t *testing.T) {
	tm := getGraph12()
	p := checkHashAlwaysReachesAndStaysCond(tm, eq0, gt10)
	assert.Nil(t, p, "bad: "+pathToString(p))
}

func TestPathStays5(t *testing.T) {
	tm := getGraph12()
	tm[11] = []stateHash{10, 11, 12}
	p := checkHashAlwaysReachesAndStaysCond(tm, eq0, gt10)
	assert.Nil(t, p, "bad: "+pathToString(p))
}

func TestPathStays6(t *testing.T) {
	tm := getGraph12()
	tm[11] = []stateHash{3}
	p := checkHashAlwaysReachesAndStaysCond(tm, eq0, gt10)
	assert.NotNil(t, p, "bad: "+pathToString(p))
}

func newCollector() *collector {
	return &collector{
		paths: []string{},
	}
}

type collector struct {
	paths []string
}

func (c *collector) fn(path []stateHash, i int) bool {
	s := pathToString(path)
	if i != -1 {
		s = fmt.Sprintf("%s:%d", s, i)
	}
	c.paths = append(c.paths, s)
	return true
}

func pathToString(path []stateHash) string {
	s := fmt.Sprint(path)
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.ReplaceAll(s, " ", ",")
	return s
}

func (c *collector) sortPaths() []string {
	sort.Strings(c.paths)
	return c.paths
}

var gt10 stateHashCondition = stateGreaterOrEqual(10)
var eq0 stateHashCondition = stateEquals(0)

func stateEquals(expected stateHash) stateHashCondition {
	return func(sh stateHash) bool {
		return sh == expected
	}
}

func stateGreaterOrEqual(expected stateHash) stateHashCondition {
	return func(sh stateHash) bool {
		return sh >= expected
	}
}

func getGraph12() map[stateHash][]stateHash {
	tm := make(map[stateHash][]stateHash)
	tm[0] = []stateHash{1, 2, 3}
	tm[1] = []stateHash{4, 5, 6}
	tm[2] = []stateHash{4, 5, 6}
	tm[3] = []stateHash{4, 5, 6}
	tm[4] = []stateHash{7, 8, 9}
	tm[5] = []stateHash{7, 8, 9}
	tm[6] = []stateHash{7, 8, 9}
	tm[7] = []stateHash{10, 11, 12}
	tm[8] = []stateHash{10, 11, 12}
	tm[9] = []stateHash{10, 11, 12}
	return tm
}
