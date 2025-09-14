package holdem_test

import (
	"encoding/json"
	"testing"

	"github.com/notnil/joker/pkg/holdem"
	"github.com/notnil/joker/pkg/jokertest"
)

func TestHand(t *testing.T) {
	dealer := jokertest.Dealer(jokertest.Deck1().Cards)
	config := holdem.Config{
		Size:     10,
		BuyInMin: 100,
		BuyInMax: 300,
		Stakes: holdem.Stakes{
			SmallBlind: 1,
			BigBlind:   2,
			Ante:       0,
		},
	}
	seats := map[int]*holdem.Player{
		0: {ID: "0", Chips: 100},
		1: {ID: "1", Chips: 100},
		2: {ID: "2", Chips: 100},
	}
	tbl, err := holdem.New(config, seats, dealer)
	if err != nil {
		t.Fatal(err)
	}
	h := tbl.NewHand()
	actions := []holdem.Action{
		{Type: holdem.Call},
		{Type: holdem.Call},
		{Type: holdem.Check},
		{Type: holdem.Bet, Chips: 2},
		{Type: holdem.Call},
		{Type: holdem.Fold},
		{Type: holdem.Check},
		{Type: holdem.Check},
		{Type: holdem.Check},
		{Type: holdem.Check},
	}
	for _, action := range actions {
		if err := h.Act(action); err != nil {
			t.Fatal(h.ActivePlayer(), h.LegalActions(), action, err, debugStr(h))
		}
	}
	if h.Results == nil {
		t.Fatal("expected results to be set")
	}
}

func debugStr(h *holdem.Hand) string {
	b, _ := json.MarshalIndent(h, "", "\t")
	return string(b)
}

