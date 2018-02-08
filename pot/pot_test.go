package pot_test

import (
	"testing"

	"github.com/notnil/joker/pot"
)

func TestAPI(t *testing.T) {
	// stacks := map[int]int{
	// 	0: 100,
	// 	1: 50,
	// 	2: 75,
	// }
	// holdem
	// pot.New(stacks, pot.Button(2), pot.Ante(1), pot.Blinds(2, 5))
	// stud
	// pot.New(stacks, pot.Ante(1)).Ante(1).SetPos(1).Bet(5).Raise(5)
}

func TestNewPot(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks)
	seat := p.SeatToAct()
	if seat == nil {
		t.Fatal("expected to find seat to act")
	}
	if seat.Pos != 0 {
		t.Fatalf("expected action to be on %d but was on %d", 0, seat.Pos)
	}
	expected := []pot.Action{pot.Fold, pot.Check, pot.Bet}
	if includes(p.PossibleActions(), expected...) == false {
		t.Fatalf("expected actions to be %v but were %v", expected, p.PossibleActions())
	}
}

func TestAnte(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, pot.Ante(2))
	chips := p.Chips()
	if chips != 6 {
		t.Fatalf("expected pot size of %d but got %d", 6, chips)
	}
	stack := p.Seats()[0].Stack
	if stack != 98 {
		t.Fatalf("expected stack size of %d but got %d", 98, stack)
	}
}

func TestBlinds(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, pot.Blinds(0, 1, 2))
	chips := p.Chips()
	if chips != 3 {
		t.Fatalf("expected pot size of %d but got %d", 3, chips)
	}
	stack := p.Seats()[2].Stack
	if stack != 98 {
		t.Fatalf("expected stack size of %d but got %d", 98, stack)
	}
}

func TestBlindsTwoSeats(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
	}
	p := pot.New(stacks, pot.Blinds(0, 1, 2))
	chips := p.Chips()
	if chips != 3 {
		t.Fatalf("expected pot size of %d but got %d", 3, chips)
	}
	stack := p.Seats()[0].Stack
	if stack != 99 {
		t.Fatalf("expected stack size of %d but got %d", 99, stack)
	}
}

func TestAllIn(t *testing.T) {
	stacks := map[int]int{
		0: 10,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, pot.Blinds(0, 1, 2))
	chips := p.Chips()
	if chips != 3 {
		t.Fatalf("expected pot size of %d but got %d", 3, chips)
	}
	stack := p.Seats()[0].Stack
	if stack != 99 {
		t.Fatalf("expected stack size of %d but got %d", 99, stack)
	}
}

func includes(actions []pot.Action, include ...pot.Action) bool {
	for _, a1 := range include {
		found := false
		for _, a2 := range actions {
			found = found || a1 == a2
		}
		if !found {
			return false
		}
	}
	return true
}
