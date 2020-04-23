package table_test

import (
	"math/rand"
	"testing"

	"github.com/notnil/joker/hand"
	"github.com/notnil/joker/table"
)

type testCase struct {
	start       *table.Table
	actions     []table.Action
	condition   func(table.State) bool
	description string
}

var (
	testCases = []testCase{
		{
			start:   threePerson100Buyin(),
			actions: nil,
			condition: func(s table.State) bool {
				return s.Seats[0].Chips == 98 && s.Seats[1].Chips == 100 && s.Seats[2].Chips == 99 && s.Active.Seat == 1
			},
			description: "initial blinds",
		},
		{
			start: threePerson100Buyin(),
			actions: []table.Action{
				{table.Raise, 5},
			},
			condition: func(s table.State) bool {
				return s.Seats[0].Chips == 98 && s.Seats[1].Chips == 93 && s.Seats[2].Chips == 99 && s.Active.Seat == 2 && s.Cost == 7
			},
			description: "preflop raise",
		},
		{
			start: threePerson100Buyin(),
			actions: []table.Action{
				{table.Raise, 5},
				{table.Call, 0},
				{table.Fold, 0},
				{table.Check, 0},
				{table.Bet, 5},
				{table.Fold, 0},
			},
			condition: func(s table.State) bool {
				return s.Seats[0].Chips == 97 && s.Seats[1].Chips == 107 && s.Seats[2].Chips == 93 && s.Active.Seat == 2 && s.Button == 2
			},
			description: "full hand 1",
		},
		{
			start: threePerson100Buyin(),
			actions: []table.Action{
				{table.Raise, 5},
				{Type: table.Fold},
				{Type: table.Fold},
			},
			condition: func(s table.State) bool {
				return s.Seats[0].Chips == 97 && s.Seats[1].Chips == 101 && s.Seats[2].Chips == 99 && s.Round == table.PreFlop
			},
			description: "preflop folds",
		},
	}
)

func TestTable(t *testing.T) {
	for _, tc := range testCases {
		tbl := tc.start
		for _, a := range tc.actions {
			if err := tbl.Act(a); err != nil {
				t.Fatal(err)
			}
		}
		if tc.condition(tbl.State()) == false {
			t.Fatalf(tc.description)
		}
	}
}

func threePerson100Buyin() *table.Table {
	src := rand.NewSource(42)
	r := rand.New(src)
	dealer := hand.NewDealer(r)
	opts := table.Options{
		Variant: table.TexasHoldem,
		Limit:   table.NoLimit,
		Stakes:  table.Stakes{SmallBlind: 1, BigBlind: 2},
		Buyin:   100,
	}
	ids := []string{"a", "b", "c"}
	return table.New(dealer, opts, ids)
}
