package bot

import (
	"math/rand"
	"testing"

	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/holdem"
)

func TestBotsOnlyLegalActions(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	styles := []Style{Nit, CallingStation, Lag}
	bots := make([]Player, 9)
	for i := range bots {
		bots[i] = New(styles[i%3].String()+"-"+string(rune('A'+i)), styles[i%3], rand.New(rand.NewSource(int64(i+1))))
	}
	tbl, err := holdem.NewTable(holdem.Config{
		Seats:  9,
		Stakes: holdem.Stakes{SmallBlind: 50, BigBlind: 100},
	}, hand.NewDealer(rng))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 9; i++ {
		if err := tbl.Sit(i, bots[i].Name(), 10000); err != nil {
			t.Fatal(err)
		}
	}
	_ = tbl.SetButton(0)

	hands := 200
	for n := 0; n < hands; n++ {
		h, err := tbl.StartHand()
		if err != nil {
			t.Fatal(err)
		}
		for !h.Done() {
			seat := h.ToAct()
			if seat < 0 {
				t.Fatal("toAct < 0 but hand not done")
			}
			v := h.View(seat)
			a := bots[seat].Act(v)
			legal := h.LegalActions()
			ok := false
			for _, tpe := range legal {
				if tpe == a.Type {
					ok = true
					break
				}
			}
			if !ok {
				t.Fatalf("hand %d seat %d played %v legal %v", n, seat, a, legal)
			}
			if err := h.Act(a); err != nil {
				t.Fatalf("hand %d seat %d act %v: %v legal %v", n, seat, a, err, legal)
			}
		}
		if err := tbl.ApplyResults(h); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 9; i++ {
			p := tbl.Player(i)
			if p != nil && p.Chips == 0 {
				_ = tbl.SetChips(i, 10000)
			}
		}
	}
}
