package deal_test

import (
	"sort"
	"testing"

	"github.com/notnil/joker/deal"
	"github.com/notnil/joker/jokertest"
	"github.com/notnil/joker/pot"
)

func TestHoldemCheckDown(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	c := deal.Config{
		Variant: deal.Holdem,
		Deck:    jokertest.Deck1(),
		Button:  0,
		Stacks:  stacks,
		Blinds:  []int{1, 2},
	}
	d := deal.New(c)
	if len(d.HoleCards()[0]) != 2 {
		t.Fatal("holdem should deal two holel cards preflop", d.HoleCards())
	}
	if len(d.Board()) != 0 {
		t.Fatal("holdem should deal no board cards preflop", d.Board())
	}
	if err := d.Action(pot.Call, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.Call, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.Check, 0); err != nil {
		t.Fatal(err)
	}
	if len(d.Board()) != 3 {
		t.Fatal("holdem should deal board cards on flop", d.Board())
	}
	// check all the way down
	for i := 0; i < 9; i++ {
		if err := d.Action(pot.Check, 0); err != nil {
			t.Fatal(err)
		}
	}
	// log.Println(d.HoleCards())
	// log.Println(d.Board())
	// log.Println(d.Payouts())
	// log.Println(d.Hands())
	if d.Payouts()[0].Pos != 2 {
		t.Fatal("player two should have won the deal")
	}
}

func TestHoldemFolds(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	c := deal.Config{
		Variant: deal.Holdem,
		Deck:    jokertest.Deck1(),
		Button:  0,
		Stacks:  stacks,
		Blinds:  []int{1, 2},
	}
	d := deal.New(c)
	if err := d.Action(pot.Fold, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.Fold, 0); err != nil {
		t.Fatal(err)
	}
	if len(d.Payouts()) == 0 {
		t.Fatal("should payout when everyone folds")
	}
}

func TestHoldemAllin(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 20,
		2: 10,
	}
	c := deal.Config{
		Variant: deal.Holdem,
		Deck:    jokertest.Deck1(),
		Button:  2,
		Stacks:  stacks,
		Blinds:  []int{1, 2},
	}
	d := deal.New(c)
	if err := d.Action(pot.Call, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.Call, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.Check, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.Check, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.Check, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.AllIn, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.AllIn, 0); err != nil {
		t.Fatal(err)
	}
	if err := d.Action(pot.Call, 0); err != nil {
		t.Fatal(err)
	}
	payouts := d.Payouts()
	sort.Slice(payouts, func(i, j int) bool {
		pi, pj := payouts[i], payouts[j]
		return pi.Pos < pj.Pos
	})
	if len(payouts) == 0 {
		t.Fatal("should payout when everyone is all in")
	}
	if payouts[0].Chips != 80 {
		t.Fatal("expected pos 0 to get 80 chips")
	}
	if payouts[1].Chips != 20 {
		t.Fatal("expected pos 1 to get 20 chips")
	}
	if payouts[2].Chips != 30 {
		t.Fatal("expected pos 2 to get 30 chips")
	}
}
