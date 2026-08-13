package play

import (
	"math/rand"
	"testing"

	"github.com/notnil/joker/pkg/holdem"
)

func TestCashSessionBotsUntilHero(t *testing.T) {
	g, err := NewCash(0, rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatal(err)
	}
	v := g.View()
	if v.Seats != 9 {
		t.Fatalf("seats = %d", v.Seats)
	}
	if v.Done {
		return
	}
	if v.ToAct != 0 && v.ToAct != -1 {
		t.Fatalf("expected hero to act or hand running out, toAct=%d", v.ToAct)
	}
}

func TestCashHumanActAndStep(t *testing.T) {
	g, err := NewCash(0, rand.New(rand.NewSource(2)))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50; i++ {
		v := g.View()
		if v.Done {
			if err := g.Step(); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if v.ToAct == g.Hero() {
			legal := v.Legal
			if len(legal) == 0 {
				t.Fatal("hero to act with no legal actions")
			}
			a, err := holdem.ParseAction(legal[0], v.MinBet)
			if err != nil {
				t.Fatal(err)
			}
			if a.Type == holdem.Bet {
				a = holdem.BetAction(v.MinBet)
			}
			if a.Type == holdem.Raise {
				a = holdem.RaiseTo(v.MinRaiseTo)
			}
			if err := g.Act(a); err != nil {
				t.Fatalf("hero %v: %v legal %v", a, err, legal)
			}
			continue
		}
		if err := g.Step(); err != nil {
			t.Fatal(err)
		}
	}
}
