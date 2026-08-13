// Package holdem implements No-Limit Texas Hold'em table mechanics:
// seating, blinds, betting, side pots, and showdown.
package holdem

import (
	"fmt"
	"slices"
)

// Pot tracks chip contributions for a single hand and which seats remain
// eligible to win them.
type Pot struct {
	contrib  map[int]int
	eligible map[int]bool
}

// SidePot is one pot (main or side) after splitting contributions by tier.
type SidePot struct {
	Amount   int
	Eligible []int
}

// NewPot returns an empty pot.
func NewPot() *Pot {
	return &Pot{
		contrib:  map[int]int{},
		eligible: map[int]bool{},
	}
}

// Add credits chips from seat. The seat becomes eligible.
func (p *Pot) Add(seat, chips int) {
	if chips < 0 {
		panic("holdem: negative pot contribution")
	}
	if chips == 0 {
		return
	}
	p.contrib[seat] += chips
	p.eligible[seat] = true
}

// Fold marks seat ineligible to win. Their chips stay in the pot.
func (p *Pot) Fold(seat int) {
	p.eligible[seat] = false
}

// Total is the sum of all contributions.
func (p *Pot) Total() int {
	n := 0
	for _, v := range p.contrib {
		n += v
	}
	return n
}

// Contribution returns chips seat has put in this hand.
func (p *Pot) Contribution(seat int) int {
	return p.contrib[seat]
}

// EligibleSeats returns seats that can still win pot chips, sorted.
func (p *Pot) EligibleSeats() []int {
	var seats []int
	for seat, ok := range p.eligible {
		if ok {
			seats = append(seats, seat)
		}
	}
	slices.Sort(seats)
	return seats
}

// SidePots splits contributions into main and side pots.
// Folded chips remain in the pot but those seats are not eligible.
// Uncalled chips (contributed only by one eligible seat) form their own
// side pot so they return to that seat.
func (p *Pot) SidePots() []SidePot {
	levels := uniqueEligibleLevels(p)
	remaining := make(map[int]int, len(p.contrib))
	for k, v := range p.contrib {
		remaining[k] = v
	}

	var pots []SidePot
	prev := 0
	for _, level := range levels {
		layer := level - prev
		sp := SidePot{}
		for seat, amt := range remaining {
			take := min(amt, layer)
			if take > 0 {
				sp.Amount += take
				remaining[seat] -= take
			}
		}
		for seat, ok := range p.eligible {
			if ok && p.contrib[seat] >= level {
				sp.Eligible = append(sp.Eligible, seat)
			}
		}
		slices.Sort(sp.Eligible)
		if sp.Amount > 0 {
			pots = append(pots, sp)
		}
		prev = level
	}

	// Uncalled remainder: chips still sitting in remaining[] belong to the
	// seats that put them in. If those seats are eligible they get them back
	// as a one-seat pot; if not, they fall into the last contested pot.
	leftoverBySeat := map[int]int{}
	leftoverTotal := 0
	for seat, amt := range remaining {
		if amt > 0 {
			leftoverBySeat[seat] = amt
			leftoverTotal += amt
		}
	}
	if leftoverTotal == 0 {
		return mergeAdjacent(pots)
	}
	var eligibleLeft []int
	ineligibleAmt := 0
	for seat, amt := range leftoverBySeat {
		if p.eligible[seat] {
			eligibleLeft = append(eligibleLeft, seat)
		} else {
			ineligibleAmt += amt
		}
	}
	slices.Sort(eligibleLeft)
	if len(eligibleLeft) > 0 {
		amt := 0
		for _, seat := range eligibleLeft {
			amt += leftoverBySeat[seat]
		}
		pots = append(pots, SidePot{Amount: amt, Eligible: eligibleLeft})
	}
	if ineligibleAmt > 0 {
		if len(pots) == 0 {
			pots = append(pots, SidePot{Amount: ineligibleAmt, Eligible: p.EligibleSeats()})
		} else {
			pots[len(pots)-1].Amount += ineligibleAmt
		}
	}
	return mergeAdjacent(pots)
}

func uniqueEligibleLevels(p *Pot) []int {
	seen := map[int]struct{}{}
	var levels []int
	for seat, ok := range p.eligible {
		if !ok {
			continue
		}
		amt := p.contrib[seat]
		if amt <= 0 {
			continue
		}
		if _, dup := seen[amt]; dup {
			continue
		}
		seen[amt] = struct{}{}
		levels = append(levels, amt)
	}
	slices.Sort(levels)
	return levels
}

func mergeAdjacent(pots []SidePot) []SidePot {
	if len(pots) < 2 {
		return pots
	}
	out := []SidePot{pots[0]}
	for _, sp := range pots[1:] {
		last := &out[len(out)-1]
		if sameSeats(last.Eligible, sp.Eligible) {
			last.Amount += sp.Amount
			continue
		}
		out = append(out, sp)
	}
	return out
}

func sameSeats(a, b []int) bool {
	return slices.Equal(a, b)
}

func splitAmount(amount, n int, oddChipOrder []int) map[int]int {
	if n == 0 || amount == 0 {
		return map[int]int{}
	}
	if len(oddChipOrder) != n {
		panic(fmt.Sprintf("holdem: odd-chip order len %d want %d", len(oddChipOrder), n))
	}
	base := amount / n
	rem := amount % n
	out := make(map[int]int, n)
	for i, seat := range oddChipOrder {
		out[seat] = base
		if i < rem {
			out[seat]++
		}
	}
	return out
}

// oddChipOrder sorts winners so the first is closest to the left of the button.
func oddChipOrder(winners []int, button, tableSize int) []int {
	out := slices.Clone(winners)
	slices.SortFunc(out, func(a, b int) int {
		return distLeftOfButton(button, a, tableSize) - distLeftOfButton(button, b, tableSize)
	})
	return out
}

func distLeftOfButton(button, seat, tableSize int) int {
	return (seat - button - 1 + tableSize) % tableSize
}
