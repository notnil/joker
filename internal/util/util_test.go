package util_test

import (
	"reflect"
	"testing"

	"github.com/loganjspears/joker/util"
)

type combo struct {
	n     int
	k     int
	combo [][]int
}

var combos = []combo{
	{n: 5, k: 5, combo: [][]int{
		[]int{0, 1, 2, 3, 4},
	}},
	{n: 3, k: 5, combo: [][]int{}},
	{n: -3, k: 5, combo: [][]int{}},
	{n: 5, k: -3, combo: [][]int{}},
	{n: 3, k: 2, combo: [][]int{
		[]int{0, 1},
		[]int{0, 2},
		[]int{1, 2},
	}},
	{n: 4, k: 2, combo: [][]int{
		[]int{0, 1},
		[]int{0, 2},
		[]int{0, 3},
		[]int{1, 2},
		[]int{1, 3},
		[]int{2, 3},
	}},
	{n: 4, k: 3, combo: [][]int{
		[]int{0, 1, 2},
		[]int{0, 1, 3},
		[]int{0, 2, 3},
		[]int{1, 2, 3},
	}},
}

func TestCombinations(t *testing.T) {
	for _, c := range combos {
		result := util.Combinations(c.n, c.k)
		if !reflect.DeepEqual(result, c.combo) {
			t.Fatalf("util.Combinations(%d, %d) => %v, want %v", c.n, c.k, result, c.combo)
		}
	}
}
