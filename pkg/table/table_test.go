package table_test

import (
	"testing"

	"github.com/notnil/joker/jokertest"
	"github.com/notnil/joker/table"
)

func TestTable(t *testing.T) {
	dealer := jokertest.Dealer(jokertest.Deck1().Cards)
	config := table.Config{
		Size:     10,
		BuyInMin: 100,
		BuyInMax: 300,
	}
	tbl, err := table.New(config, nil, dealer)
	if err != nil {
		t.Fatal(err)
	}
	p1 := &table.Player{
		ID:    "1",
		Chips: 200,
	}
	p2 := &table.Player{
		ID:    "2",
		Chips: 200,
	}
	if err := tbl.Sit(0, p1); err != nil {
		t.Fatal(err)
	}
	if err := tbl.Sit(1, p2); err != nil {
		t.Fatal(err)
	}
	next0 := tbl.Next(0)
	if next0 != 1 {
		t.Fatalf("expected the next seat of %d to be %d but got %d", 0, 1, next0)
	}
	next1 := tbl.Next(1)
	if next1 != 0 {
		t.Fatalf("expected the next seat of %d to be %d but got %d", 1, 0, next1)
	}
}
