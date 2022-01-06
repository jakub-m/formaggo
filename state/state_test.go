package state

import (
	"testing"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	tcs := []struct {
		a, b  interface{}
		equal bool
	}{
		{
			123,
			123,
			true,
		},
		{
			struct{ A, B int }{100, 200},
			struct{ A, B int }{100, 200},
			true,
		},
		{
			struct{ A, B int }{100, 200},
			struct{ A, B int }{200, 100},
			false,
		},
		{
			struct{ x int }{100},
			struct{ y int }{200},
			true, // private fields are ignored!
		},
		{
			struct{ Slice []int }{[]int{1, 3}},
			struct{ Slice []int }{[]int{3, 1}},
			false,
		},
		{
			struct {
				Slice []int `hash:"set"`
			}{[]int{1, 2}},
			struct {
				Slice []int `hash:"set"`
			}{[]int{2, 1}},
			true,
		},
		{
			struct{ s []string }{[]string{"foo", "bar"}},
			struct{ s []string }{[]string{"foo", "bar"}},
			true,
		},
	}
	for _, tc := range tcs {
		h1, err := hashstructure.Hash(tc.a, hashstructure.FormatV2, nil)
		assert.NoError(t, err)
		h2, err := hashstructure.Hash(tc.b, hashstructure.FormatV2, nil)
		assert.NoError(t, err)
		if tc.equal {
			assert.Equal(t, h1, h2, "should be equal: %+v, %+v", tc.a, tc.b)
		} else {
			assert.NotEqual(t, h1, h2, "should not be equal: %+v, %+v", tc.a, tc.b)
		}
	}
}
