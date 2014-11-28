package table

import (
	"encoding/json"
	"testing"

	"github.com/SyntropyDev/joker/hand"
	"github.com/SyntropyDev/joker/jokertest"
)

func TestPotJSON(t *testing.T) {
	pot := newPot(3)
	pot.contribute(0, 1)

	b, err := json.Marshal(pot)
	if err != nil {
		t.Fatal(err)
	}

	// unmarshal from json
	p := &Pot{}
	if err := json.Unmarshal(b, p); err != nil {
		t.Fatal(err)
	}
	if p.Chips() != 1 {
		t.Fatal("pot json deserialization unsuccessful")
	}
}

func TestHighPot(t *testing.T) {
	t.Parallel()

	pot := newPot(3)
	pot.contribute(0, 5)
	pot.contribute(1, 10)
	pot.contribute(2, 15)

	seatToHoleCards := map[int][]*HoleCard{
		0: []*HoleCard{
			newHoleCard(hand.AceSpades, Concealed),
			newHoleCard(hand.AceHearts, Concealed),
		},
		1: []*HoleCard{
			newHoleCard(hand.QueenSpades, Concealed),
			newHoleCard(hand.QueenHearts, Concealed),
		},
		2: []*HoleCard{
			newHoleCard(hand.KingSpades, Concealed),
			newHoleCard(hand.KingHearts, Concealed),
		},
	}

	board := jokertest.Cards("Ad", "Kd", "Qd", "2d", "2h")
	hands := newHands(seatToHoleCards, board, Holdem.getGameType().highHand)
	payout := pot.payout(hands, tableHands(map[int]*hand.Hand{}), winHigh, 0)

	for seat, results := range payout {
		switch seat {
		case 0:
			if len(results) != 1 {
				t.Fatal("seat 0 should win one pot")
			}
		case 1:
			if len(results) != 0 {
				t.Fatal("seat 1 should win no pots")
			}
		case 2:
			if len(results) != 2 {
				t.Fatal("seat 2 should win two pots")
			}
		}
	}
}

func TestHighLowPot(t *testing.T) {
	t.Parallel()

	pot := newPot(3)
	pot.contribute(0, 5)
	pot.contribute(1, 5)
	pot.contribute(2, 5)

	seatToHoleCards := map[int][]*HoleCard{
		0: []*HoleCard{
			newHoleCard(hand.AceHearts, Concealed),
			newHoleCard(hand.TwoClubs, Concealed),
			newHoleCard(hand.SevenDiamonds, Concealed),
			newHoleCard(hand.KingHearts, Concealed),
		},
		1: []*HoleCard{
			newHoleCard(hand.AceDiamonds, Concealed),
			newHoleCard(hand.FourClubs, Concealed),
			newHoleCard(hand.ThreeDiamonds, Concealed),
			newHoleCard(hand.SixSpades, Concealed),
		},
		2: []*HoleCard{
			newHoleCard(hand.AceSpades, Concealed),
			newHoleCard(hand.TwoHearts, Concealed),
			newHoleCard(hand.JackDiamonds, Concealed),
			newHoleCard(hand.JackClubs, Concealed),
		},
	}

	board := jokertest.Cards("7s", "Kd", "8h", "Jh", "5c")
	highHands := newHands(seatToHoleCards, board, OmahaHiLo.getGameType().highHand)
	lowHands := newHands(seatToHoleCards, board, OmahaHiLo.getGameType().lowHand)
	payout := pot.payout(highHands, lowHands, winHighLow, 0)

	for seat, results := range payout {
		switch seat {
		case 0:
			if len(results) != 1 && total(results) != 4 {
				t.Fatal("seat 0 should win 4 chips")
			}
		case 1:
			if len(results) != 1 && total(results) != 8 {
				t.Fatal("seat 1 should win 8 chips")
			}
		case 2:
			if len(results) != 1 && total(results) != 3 {
				t.Fatal("seat 2 should 3 chips")
			}
		}
	}
}

func total(results []*PotResult) int {
	chips := 0
	for _, r := range results {
		chips += r.Chips
	}
	return chips
}
