package deal

import (
	"testing"

	"github.com/notnil/joker/hand"
	"github.com/notnil/joker/jokertest"
)

type evalTest struct {
	hole  []hand.Card
	board []hand.Card
	eval  handEvaluator
	hand  *hand.Hand
}

var (
	evalTests = []evalTest{
		{
			hole:  jokertest.Cards("As", "Ks", "4c", "2d"),
			board: jokertest.Cards("Ac", "Ad", "Qd", "Jd", "5d"),
			eval:  omahaHiHandEvaluator{},
			hand:  hand.New(jokertest.Cards("Ac", "Ad", "As", "Ks", "Qd")),
		},
		{
			hole:  jokertest.Cards("As", "Ks", "4c", "2d"),
			board: jokertest.Cards("7c", "5d", "3s", "Jd", "5d"),
			eval:  omahaLowHandEvaluator{},
			hand:  hand.New(jokertest.Cards("7c", "5d", "3s", "2d", "As"), hand.AceToFiveLow),
		},
		{
			hole:  jokertest.Cards("As", "Ks", "4c", "2d"),
			board: jokertest.Cards("Jc", "5d", "3s", "Jd", "5d"),
			eval:  omahaLowHandEvaluator{},
			hand:  nil,
		},
		{
			hole:  jokertest.Cards("As", "Ks", "4c", "2d", "3s", "Jd", "5d"),
			board: jokertest.Cards(),
			eval:  studLow8HandEvaluator{},
			hand:  hand.New(jokertest.Cards("5d", "4c", "3s", "2d", "As"), hand.AceToFiveLow),
		},
		{
			hole:  jokertest.Cards("As", "Ks", "4c", "2d", "3s", "Jd", "9d"),
			board: jokertest.Cards(),
			eval:  studLow8HandEvaluator{},
			hand:  nil,
		},
	}
)

func TestEvals(t *testing.T) {
	for _, e := range evalTests {
		h := e.eval.EvaluateHand(e.hole, e.board)
		if h == nil {
			if e.hand != nil {
				t.Fatalf("expected %s but got %s", e.hand, h)
			}
		} else if h.CompareTo(e.hand) != 0 {
			t.Fatalf("expected %s but got %s", e.hand, h)
		}
	}
}
