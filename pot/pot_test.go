package pot_test

import (
	"testing"

	"github.com/notnil/joker/pot"
)

func TestNewPot(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, 0)
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
	p := pot.New(stacks, 0, pot.Ante(2))
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
	p := pot.New(stacks, 0, pot.Blinds([]int{1, 2}))
	chips := p.Chips()
	if chips != 3 {
		t.Fatalf("expected pot size of %d but got %d", 3, chips)
	}
	stack := p.Seats()[2].Stack
	if stack != 98 {
		t.Fatalf("expected stack size of %d but got %d", 98, stack)
	}
	if p.SeatToAct().Pos != 0 {
		t.Fatalf("expected starting pos of %d but got %d", 0, p.SeatToAct().Pos)
	}
}

func TestBringIn(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, 0, pot.BringIn(0, 2))
	if cost := p.Cost(); cost != 2 {
		t.Fatalf("expected cost of %d but got %d", 2, cost)
	}
	expected := []pot.Action{pot.Call, pot.Raise}
	if includes(p.PossibleActions(), expected...) == false {
		t.Fatalf("expected actions to be %v but were %v", expected, p.PossibleActions())
	}
	if err := p.Call(); err != nil {
		t.Fatal(err)
	}
}

func TestBlindsTwoSeats(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
	}
	p := pot.New(stacks, 0, pot.Blinds([]int{1, 2}))
	chips := p.Chips()
	if chips != 3 {
		t.Fatalf("expected pot size of %d but got %d", 3, chips)
	}
	stack := p.Seats()[0].Stack
	if stack != 99 {
		t.Fatalf("expected stack size of %d but got %d", 99, stack)
	}
	if p.SeatToAct().Pos != 0 {
		t.Fatalf("expected starting pos of %d but got %d", 0, p.SeatToAct().Pos)
	}
}

func TestAllIn(t *testing.T) {
	stacks := map[int]int{
		0: 10,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, 0, pot.Blinds([]int{1, 2}))
	if err := p.AllIn(); err != nil {
		t.Fatal(err)
	}
	if p.Chips() != 13 {
		t.Fatalf("expected pot size of %d but got %d", 13, p.Chips())
	}
	if p.Cost() != 9 {
		t.Fatalf("expected pot cost of %d but got %d", 9, p.Cost())
	}
}

func TestBasicHoldem(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, 0, pot.Blinds([]int{1, 2}))
	if err := p.Call(); err != nil {
		t.Fatal(err)
	}
	if err := p.Call(); err != nil {
		t.Fatal(err)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	if p.SeatToAct() != nil {
		t.Fatal("failed to recognize end of betting")
	}
	p.NextRound()
	if p.SeatToAct() == nil || p.SeatToAct().Pos != 1 {
		t.Fatal("failed reset button")
	}
	if err := p.Bet(5); err != nil {
		t.Fatal(err)
	}
	if err := p.Raise(5); err != nil {
		t.Fatal(err)
	}
	if err := p.Call(); err != nil {
		t.Fatal(err)
	}
	if err := p.Call(); err != nil {
		t.Fatal(err)
	}
	if p.SeatToAct() != nil {
		t.Fatal("failed to recognize end of betting")
	}
	p.NextRound()
	if err := p.AllIn(); err != nil {
		t.Fatal(err)
	}
	if err := p.Call(); err != nil {
		t.Fatal(err)
	}
	if err := p.Call(); err != nil {
		t.Fatal(err)
	}
	if p.SeatToAct() != nil {
		t.Fatal("failed to recognize end of betting")
	}
	p.NextRound()
	if p.SeatToAct() != nil {
		t.Fatal("failed to recognize that no betting can occur in round")
	}
}

func TestPayout(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, 0, pot.Blinds([]int{1, 2}))
	p.Call()
	p.Call()
	hi := [][]int{[]int{1}}
	payouts := p.Payout(hi, nil)
	if len(payouts) != 1 {
		t.Fatal("invalid number of payouts")
	}
	if payouts[0].Pos != 1 {
		t.Fatal("wrong player won")
	}
	if payouts[0].Share != pot.WonHigh {
		t.Fatal("player should have won high")
	}
}

func TestUncontested(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, 0, pot.Blinds([]int{1, 2}))
	if err := p.Fold(); err != nil {
		t.Fatal(err)
	}
	if err := p.Fold(); err != nil {
		t.Fatal(err)
	}
	if seat := p.SeatToAct(); seat != nil {
		t.Fatal("should not have seat to act when everyone else folds")
	}
	payout := p.Uncontested()
	if payout == nil {
		t.Fatal("should have payout")
	}
	if payout.Pos != 2 {
		t.Fatal("pos two should have won")
	}
	if payout.Chips != 3 {
		t.Fatal("pos two should have won 3 chips")
	}
	if payout.Share != pot.WonUncontested {
		t.Fatal("pos two should have won uncontested")
	}
}

func TestSplitPayout(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	p := pot.New(stacks, 0, pot.Blinds([]int{1, 2}))
	p.Call()
	p.Call()
	hi := [][]int{[]int{1, 2}}
	payouts := p.Payout(hi, nil)
	if len(payouts) != 2 {
		t.Fatal("invalid number of payouts")
	}
	if payouts[0].Share != pot.SplitHigh {
		t.Fatal("player should have split high")
	}
	if payouts[0].Chips != 3 {
		t.Fatal("player should have split high")
	}
}

func TestSplitLowPayout(t *testing.T) {
	stacks := map[int]int{
		0: 10,
		1: 20,
		2: 30,
	}
	p := pot.New(stacks, 0, pot.Blinds([]int{1, 2}))
	p.AllIn()
	p.AllIn()
	p.Call()
	hi := [][]int{[]int{0, 2}}
	low := [][]int{[]int{0}}
	payouts := p.Payout(hi, low)
	if len(payouts) != 4 {
		t.Fatal("invalid number of payouts")
	}
	// TODO equal this
	// [{"Pos":0,"Chips":8,"Share":"SplitHigh"} {"Pos":2,"Chips":7,"Share":"SplitHigh"} {"Pos":0,"Chips":15,"Share":"WonLow"} {"Pos":2,"Chips":20,"Share":"WonHigh"}]
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
