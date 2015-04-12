package game

// import (
// 	"encoding/json"
// 	"testing"

// 	"github.com/loganjspears/joker/hand"
// 	"github.com/loganjspears/joker/jokertest"
// 	"github.com/loganjspears/joker/pot"
// 	"github.com/loganjspears/joker/util"
// )

// var (
// 	holdemFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
// 		cards := append(board, holeCards...)
// 		return hand.New(cards)
// 	}

// 	omahaHiFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
// 		opts := func(c *hand.Config) {}
// 		hands := omahaHands(holeCards, board, opts)
// 		hands = hand.Sort(hand.SortingHigh, hand.DESC, hands...)
// 		return hands[0]
// 	}

// 	omahaLoFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
// 		hands := omahaHands(holeCards, board, hand.AceToFiveLow)
// 		hands = hand.Sort(hand.SortingLow, hand.DESC, hands...)
// 		if hands[0].CompareTo(eightOrBetter) <= 0 {
// 			return hands[0]
// 		}
// 		return nil
// 	}

// 	eightOrBetter = hand.New([]*hand.Card{
// 		hand.EightSpades,
// 		hand.SevenSpades,
// 		hand.SixSpades,
// 		hand.FiveSpades,
// 		hand.FourSpades,
// 	}, hand.AceToFiveLow)
// )

// func TestPotJSON(t *testing.T) {
// 	t.Parallel()

// 	p := pot.New(3)
// 	p.Contribute(0, 1)

// 	b, err := json.Marshal(p)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// unmarshal from json
// 	p = &pot.Pot{}
// 	if err := json.Unmarshal(b, p); err != nil {
// 		t.Fatal(err)
// 	}
// 	if p.Chips() != 1 {
// 		t.Errorf("after json roundtrip pot.Chips() = %v; want %v", p.Chips(), 1)
// 	}
// }

// func TestHighPot(t *testing.T) {
// 	t.Parallel()

// 	p := pot.New(3)
// 	p.Contribute(0, 5)
// 	p.Contribute(1, 10)
// 	p.Contribute(2, 15)

// 	seatToHoleCards := map[int][]*hand.Card{
// 		0: []*hand.Card{
// 			hand.AceSpades,
// 			hand.AceHearts,
// 		},
// 		1: []*hand.Card{
// 			hand.QueenSpades,
// 			hand.QueenHearts,
// 		},
// 		2: []*hand.Card{
// 			hand.KingSpades,
// 			hand.KingHearts,
// 		},
// 	}

// 	board := jokertest.Cards("Ad", "Kd", "Qd", "2d", "2h")
// 	hands := pot.NewHands(seatToHoleCards, board, holdemFunc)
// 	payout := p.Payout(hands, nil, hand.SortingHigh, 0)

// 	for seat, results := range payout {
// 		switch seat {
// 		case 0:
// 			if len(results) != 1 {
// 				t.Fatal("seat 0 should win one pot")
// 			}
// 		case 1:
// 			if len(results) != 0 {
// 				t.Fatal("seat 1 should win no pots")
// 			}
// 		case 2:
// 			if len(results) != 2 {
// 				t.Fatal("seat 2 should win two pots")
// 			}
// 		}
// 	}
// }

// func TestHighLowPot(t *testing.T) {
// 	t.Parallel()

// 	p := pot.New(3)
// 	p.Contribute(0, 5)
// 	p.Contribute(1, 5)
// 	p.Contribute(2, 5)

// 	seatToHoleCards := map[int][]*hand.Card{
// 		0: []*hand.Card{
// 			hand.AceHearts,
// 			hand.TwoClubs,
// 			hand.SevenDiamonds,
// 			hand.KingHearts,
// 		},
// 		1: []*hand.Card{
// 			hand.AceDiamonds,
// 			hand.FourClubs,
// 			hand.ThreeDiamonds,
// 			hand.SixSpades,
// 		},
// 		2: []*hand.Card{
// 			hand.AceSpades,
// 			hand.TwoHearts,
// 			hand.JackDiamonds,
// 			hand.JackClubs,
// 		},
// 	}

// 	board := jokertest.Cards("7s", "Kd", "8h", "Jh", "5c")
// 	highHands := pot.NewHands(seatToHoleCards, board, omahaHiFunc)
// 	lowHands := pot.NewHands(seatToHoleCards, board, omahaLoFunc)
// 	payout := p.Payout(highHands, lowHands, hand.SortingHigh, 0)

// 	if len(payout) < 3 {
// 		t.Errorf("pot.Payout() should have 3 results")
// 	}

// 	for seat, results := range payout {
// 		switch seat {
// 		case 0:
// 			if len(results) != 1 && total(results) != 4 {
// 				t.Errorf("seat 0 should win 4 chips")
// 			}
// 		case 1:
// 			if len(results) != 1 && total(results) != 8 {
// 				t.Errorf("seat 1 should win 8 chips")
// 			}
// 		case 2:
// 			if len(results) != 1 && total(results) != 3 {
// 				t.Errorf("seat 2 should 3 chips")
// 			}
// 		}
// 	}
// }

// func total(results []*pot.Result) int {
// 	chips := 0
// 	for _, r := range results {
// 		chips += r.Chips
// 	}
// 	return chips
// }

// func omahaHands(holeCards []*hand.Card, board []*hand.Card, opts func(*hand.Config)) []*hand.Hand {
// 	hands := []*hand.Hand{}
// 	for _, indexes := range util.Combinations(4, 2) {
// 		selected := []*hand.Card{}
// 		for _, i := range indexes {
// 			selected = append(selected, holeCards[i])
// 		}
// 		cards := append(board, selected...)
// 		hands = append(hands, hand.New(cards, opts))
// 	}
// 	return hands
// }
