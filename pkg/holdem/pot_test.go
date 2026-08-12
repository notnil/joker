package holdem

import (
	"testing"
)

func TestSidePotsClassic(t *testing.T) {
	p := NewPot()
	p.Add(0, 5)
	p.Add(1, 10)
	p.Add(2, 10)
	p.Add(3, 7)
	p.Fold(3)
	pots := p.SidePots()
	if p.Total() != 32 {
		t.Fatalf("total = %d", p.Total())
	}
	if len(pots) != 2 {
		t.Fatalf("pots = %+v", pots)
	}
	if pots[0].Amount != 20 || !sameSeats(pots[0].Eligible, []int{0, 1, 2}) {
		t.Fatalf("main = %+v", pots[0])
	}
	if pots[1].Amount != 12 || !sameSeats(pots[1].Eligible, []int{1, 2}) {
		t.Fatalf("side = %+v", pots[1])
	}
}

func TestSidePotsUncalledBet(t *testing.T) {
	p := NewPot()
	p.Add(0, 50)
	p.Add(1, 100)
	p.Fold(0)
	pots := p.SidePots()
	if len(pots) != 1 {
		t.Fatalf("pots = %+v", pots)
	}
	if pots[0].Amount != 150 || !sameSeats(pots[0].Eligible, []int{1}) {
		t.Fatalf("pot = %+v", pots[0])
	}
}

func TestSidePotsUncalledMerged(t *testing.T) {
	p := NewPot()
	p.Add(0, 50)
	p.Add(1, 100)
	p.Fold(0)
	pots := p.SidePots()
	sum := 0
	for _, sp := range pots {
		sum += sp.Amount
		if !sameSeats(sp.Eligible, []int{1}) {
			t.Fatalf("eligible = %v", sp.Eligible)
		}
	}
	if sum != 150 {
		t.Fatalf("sum = %d", sum)
	}
}

func TestSidePotsThreeWayAllIn(t *testing.T) {
	p := NewPot()
	p.Add(0, 30)
	p.Add(1, 50)
	p.Add(2, 100)
	pots := p.SidePots()
	if len(pots) != 3 {
		t.Fatalf("pots = %+v", pots)
	}
	if pots[0].Amount != 90 || !sameSeats(pots[0].Eligible, []int{0, 1, 2}) {
		t.Fatalf("main = %+v", pots[0])
	}
	if pots[1].Amount != 40 || !sameSeats(pots[1].Eligible, []int{1, 2}) {
		t.Fatalf("side = %+v", pots[1])
	}
	if pots[2].Amount != 50 || !sameSeats(pots[2].Eligible, []int{2}) {
		t.Fatalf("uncalled = %+v", pots[2])
	}
}

func TestSplitAmountOddChip(t *testing.T) {
	got := splitAmount(100, 3, []int{2, 0, 1})
	if got[2] != 34 || got[0] != 33 || got[1] != 33 {
		t.Fatalf("got %v", got)
	}
}

func TestOddChipOrder(t *testing.T) {
	// button 0, left is 1 then 2 then 3
	order := oddChipOrder([]int{3, 1, 2}, 0, 9)
	if !sameSeats(order, []int{1, 2, 3}) {
		t.Fatalf("order = %v", order)
	}
}
