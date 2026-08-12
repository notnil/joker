package tournament

import (
	"testing"

	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/holdem"
	"github.com/notnil/joker/pkg/jokertest"
)

func TestICMClassicThreeWay(t *testing.T) {
	got := ICM([]int{500, 300, 200}, []int{50, 30, 20})
	if sumInt(got) != 100 {
		t.Fatalf("sum = %d, got %v", sumInt(got), got)
	}
	// Published Malmuth–Harville: ~38.39, 32.75, 28.75 → 38, 33, 29
	want := []int{38, 33, 29}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ICM = %v, want %v", got, want)
		}
	}
}

func TestICMEqualStacks(t *testing.T) {
	got := ICM([]int{100, 100, 100}, []int{50, 30, 20})
	if sumInt(got) != 100 {
		t.Fatalf("sum = %d %v", sumInt(got), got)
	}
	for _, v := range got {
		if v < 33 || v > 34 {
			t.Fatalf("equal stacks should be ~33 each, got %v", got)
		}
	}
}

func TestICMSinglePlayer(t *testing.T) {
	got := ICM([]int{1000, 0, 0}, []int{50, 30, 20})
	if got[0] != 100 || got[1] != 0 || got[2] != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestSNGHeadsUpFoldWin(t *testing.T) {
	d := jokertest.Dealer(append(jokertest.Cards("As", "Ks", "2h", "3h"), hand.Cards()...))
	sng, err := NewSNG(Config{
		Seats:         2,
		StartingStack: 1000,
		Payouts:       []int{150, 50},
		Blinds:        []BlindLevel{{SmallBlind: 50, BigBlind: 100, Hands: 10}},
	}, []Entrant{{ID: "a", Seat: 0}, {ID: "b", Seat: 1}}, d)
	if err != nil {
		t.Fatal(err)
	}
	h, err := sng.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Act(holdem.FoldAction()); err != nil {
		t.Fatal(err)
	}
	if err := sng.FinishHand(h); err != nil {
		t.Fatal(err)
	}
	if sng.Done() {
		t.Fatal("should still be playing")
	}
	if sng.Table().Player(0).Chips != 950 || sng.Table().Player(1).Chips != 1050 {
		t.Fatalf("stacks %d %d", sng.Table().Player(0).Chips, sng.Table().Player(1).Chips)
	}
}

func TestSNGBustAndPayout(t *testing.T) {
	// Short stack shoves and loses; tournament ends.
	d := jokertest.Dealer(jokertest.Cards(
		"2s", "3s", // SB/button seat 0
		"As", "Ah", // BB seat 1
		"7c", "8d", "9s", "Jh", "Qc",
	))
	sng, err := NewSNG(Config{
		Seats:         2,
		StartingStack: 300,
		Payouts:       []int{100, 50},
		Blinds: []BlindLevel{
			{SmallBlind: 50, BigBlind: 100, Hands: 1},
			{SmallBlind: 100, BigBlind: 200, Hands: 10},
		},
	}, []Entrant{{ID: "short", Seat: 0}, {ID: "big", Seat: 1}}, d)
	if err != nil {
		t.Fatal(err)
	}
	h, err := sng.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Act(holdem.AllInAction()); err != nil {
		t.Fatal(err)
	}
	if err := h.Act(holdem.CallAction()); err != nil {
		t.Fatal(err)
	}
	if !h.Done() {
		t.Fatal("expected showdown")
	}
	if err := sng.FinishHand(h); err != nil {
		t.Fatal(err)
	}
	if !sng.Done() {
		t.Fatal("tournament should be over")
	}
	st := sng.Standings()
	if len(st) != 2 || st[0].Place != 1 || st[0].ID != "big" {
		t.Fatalf("standings = %+v", st)
	}
	if st[0].Prize != 100 || st[1].Prize != 50 {
		t.Fatalf("prizes %+v", st)
	}
}

func TestBlindLevelAdvance(t *testing.T) {
	d := jokertest.Dealer(hand.Cards())
	sng, err := NewSNG(Config{
		Seats:         2,
		StartingStack: 5000,
		Payouts:       []int{1},
		Blinds: []BlindLevel{
			{SmallBlind: 25, BigBlind: 50, Hands: 1},
			{SmallBlind: 50, BigBlind: 100, Hands: 10},
		},
	}, []Entrant{{ID: "a", Seat: 0}, {ID: "b", Seat: 1}}, d)
	if err != nil {
		t.Fatal(err)
	}
	h, err := sng.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	if h.MinBet() != 50 {
		t.Fatalf("bb = %d", h.MinBet())
	}
	if err := h.Act(holdem.FoldAction()); err != nil {
		t.Fatal(err)
	}
	if err := sng.FinishHand(h); err != nil {
		t.Fatal(err)
	}
	if sng.Level().BigBlind != 100 {
		t.Fatalf("level = %+v", sng.Level())
	}
}
