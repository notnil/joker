package table

import (
	"encoding/json"
	"sort"
)

type Pot struct {
	contributions map[int]int
	eligible      map[int]bool
}

func NewPot(contributions map[int]int) *Pot {
	cp := map[int]int{}
	eligible := map[int]bool{}
	for k, v := range contributions {
		cp[k] = v
		eligible[k] = true
	}
	return &Pot{
		contributions: cp,
		eligible:      eligible,
	}
}

func (p *Pot) Add(seat int, chips int) {
	p.contributions[seat] = p.contributions[seat] + chips
	p.eligible[seat] = true
}

func (p *Pot) Remove(seat int) {
	// TODO check if highest contribution and don't allow
	p.eligible[seat] = false
}

func (p *Pot) Total() int {
	total := 0
	for _, v := range p.contributions {
		total += v
	}
	return total
}

func (p *Pot) Cost() int {
	cost := 0
	for _, v := range p.contributions {
		cost = max(cost, v)
	}
	return cost
}

func (p *Pot) Contribution(seat int) int {
	return p.contributions[seat]
}

func (p *Pot) Eligible() []int {
	a := []int{}
	for k := range p.eligible {
		a = append(a, k)
	}
	return a
}

func (p *Pot) Owe(seat int) int {
	return p.Cost() - p.Contribution(seat)
}

func (p *Pot) Split() []*Pot {
	unique := map[int]struct{}{}
	for k, v := range p.contributions {
		if p.eligible[k] {
			unique[v] = struct{}{}
		}
	}
	amounts := []int{}
	for v := range unique {
		amounts = append(amounts, v)
	}
	sort.IntSlice(amounts).Sort()

	pots := []*Pot{}
	cp := p.Copy()
	for i, chips := range amounts {
		last := 0
		if i != 0 {
			last = amounts[i-1]
		}
		pot := NewPot(nil)
		for seat, contrib := range cp.contributions {
			amount := min(contrib, chips-last)
			if amount > 0 {
				pot.Add(seat, amount)
				cp.contributions[seat] -= amount
			}
		}
		pots = append(pots, pot)
	}
	return pots
}

func (p *Pot) Copy() *Pot {
	contributions := map[int]int{}
	eligible := map[int]bool{}
	for k, v := range p.contributions {
		contributions[k] = v
	}
	for k, v := range p.eligible {
		eligible[k] = v
	}
	return &Pot{contributions: contributions, eligible: eligible}
}

type potJSON struct {
	Contributions map[int]int  `json:"contributions"`
	Eligible      map[int]bool `json:"eligible"`
}

func (p *Pot) MarshalJSON() ([]byte, error) {
	js := &potJSON{
		Contributions: p.contributions,
		Eligible:      p.eligible,
	}
	return json.Marshal(js)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}
