package deal_test

import (
	"testing"

	"github.com/notnil/joker/pot"

	"github.com/notnil/joker/deal"
	"github.com/notnil/joker/jokertest"
)

func TestHoldemDeal(t *testing.T) {
	stacks := map[int]int{
		0: 100,
		1: 100,
		2: 100,
	}
	deck1 := jokertest.Deck1()
	d := deal.New(deal.Holdem, deck1, stacks, 0)
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
