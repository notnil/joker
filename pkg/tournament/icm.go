package tournament

import (
	"cmp"
	"math"
	"slices"
)

// ICM computes Malmuth–Harville Independent Chip Model equities for a
// single table (up to 10 players). Returned values are integer minor units
// that sum to the sum of payouts. Zero stacks are treated as already
// eliminated (equity 0).
func ICM(stacks, payouts []int) []int {
	n := len(stacks)
	out := make([]int, n)
	if n == 0 {
		return out
	}
	var live []int
	for i, s := range stacks {
		if s > 0 {
			live = append(live, i)
		}
	}
	if len(live) == 0 {
		return out
	}
	if len(live) == 1 {
		if len(payouts) > 0 {
			out[live[0]] = sumInt(payouts)
		}
		return out
	}

	eq := icmEquity(stacks, payouts)
	return roundToSum(eq, sumInt(payouts))
}

func icmEquity(stacks, payouts []int) []float64 {
	n := len(stacks)
	memo := map[uint32][]float64{}
	var all uint32
	for i, s := range stacks {
		if s > 0 {
			all |= 1 << i
		}
	}
	var rec func(mask uint32) []float64
	rec = func(mask uint32) []float64 {
		if cached, ok := memo[mask]; ok {
			return cached
		}
		eq := make([]float64, n)
		remaining := bits(mask, n)
		if len(remaining) == 0 {
			memo[mask] = eq
			return eq
		}
		place := popcount(all) - popcount(mask)
		prize := 0.0
		if place >= 0 && place < len(payouts) {
			prize = float64(payouts[place])
		}
		if len(remaining) == 1 {
			eq[remaining[0]] = prize
			memo[mask] = eq
			return eq
		}
		total := 0
		for _, i := range remaining {
			total += stacks[i]
		}
		for _, j := range remaining {
			p := float64(stacks[j]) / float64(total)
			sub := rec(mask &^ (1 << j))
			eq[j] += p * prize
			for _, i := range remaining {
				if i == j {
					continue
				}
				eq[i] += p * sub[i]
			}
		}
		memo[mask] = eq
		return eq
	}
	return rec(all)
}

func bits(mask uint32, n int) []int {
	var out []int
	for i := 0; i < n; i++ {
		if mask&(1<<i) != 0 {
			out = append(out, i)
		}
	}
	return out
}

func popcount(m uint32) int {
	n := 0
	for m != 0 {
		n += int(m & 1)
		m >>= 1
	}
	return n
}

func sumInt(a []int) int {
	n := 0
	for _, v := range a {
		n += v
	}
	return n
}

func roundToSum(eq []float64, total int) []int {
	n := len(eq)
	out := make([]int, n)
	if total == 0 {
		return out
	}
	type frac struct {
		i int
		f float64
	}
	sum := 0
	fracs := make([]frac, n)
	for i, v := range eq {
		floor := int(math.Floor(v + 1e-9))
		out[i] = floor
		sum += floor
		fracs[i] = frac{i: i, f: v - float64(floor)}
	}
	slices.SortFunc(fracs, func(a, b frac) int {
		if a.f != b.f {
			return cmp.Compare(b.f, a.f)
		}
		return a.i - b.i
	})
	for k := 0; k < total-sum && k < len(fracs); k++ {
		out[fracs[k].i]++
	}
	return out
}
