package table_test

import (
	"encoding/json"
	"testing"

	"github.com/notnil/joker/jokertest"
	"github.com/notnil/joker/table"
)

func TestHand(t *testing.T) {
	dealer := jokertest.Dealer(jokertest.Deck1().Cards)
	config := table.Config{
		Size:     10,
		BuyInMin: 100,
		BuyInMax: 300,
		Stakes: table.Stakes{
			SmallBlind: 1,
			BigBlind:   2,
			Ante:       0,
		},
	}
	seats := map[int]*table.Player{
		0: {ID: "0", Chips: 100},
		1: {ID: "1", Chips: 100},
		2: {ID: "2", Chips: 100},
	}
	tbl, err := table.New(config, seats, dealer)
	if err != nil {
		t.Fatal(err)
	}
	h := tbl.NewHand()
	actions := []table.Action{
		{Type: table.Call},
		{Type: table.Call},
		{Type: table.Check},
		{Type: table.Bet, Chips: 2},
		{Type: table.Call},
		{Type: table.Fold},
		{Type: table.Check},
		{Type: table.Check},
		{Type: table.Check},
		{Type: table.Check},
	}
	for _, action := range actions {
		if err := h.Act(action); err != nil {
			t.Fatal(h.ActivePlayer(), h.LegalActions(), action, err, debugStr(h))
		}
	}
	t.Fatal(h.Results)
}

func debugStr(h *table.Hand) string {
	b, _ := json.MarshalIndent(h, "", "\t")
	return string(b)
}
