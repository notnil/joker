package game

// import (
// 	"testing"

// 	"github.com/loganjspears/joker/hand"
// 	"github.com/loganjspears/joker/jokertest"
// 	"github.com/loganjspears/joker/pot"
// )

// var tests = []struct {
// 	G              Game
// 	Cards          []*hand.Card
// 	NumOfHoleCards int
// 	NumOfBoard     int
// 	HighRanking    hand.Ranking
// 	LowRanking     hand.Ranking
// }{
// 	{
// 		G:              Holdem,
// 		Cards:          jokertest.Cards("Qh", "Ks", "4s", "3d", "4s", "8h", "2c", "Ah", "Kh"),
// 		NumOfHoleCards: 2,
// 		NumOfBoard:     5,
// 		HighRanking:    hand.Pair,
// 	},
// 	{
// 		G:              OmahaHiLo,
// 		Cards:          jokertest.Cards("Qh", "Ks", "4s", "3d", "4s", "8h", "2c", "Ah", "Kh"),
// 		NumOfHoleCards: 4,
// 		NumOfBoard:     5,
// 		HighRanking:    hand.TwoPair,
// 		LowRanking:     hand.HighCard,
// 	},
// 	{
// 		G:              StudHiLo,
// 		Cards:          jokertest.Cards("Ah", "Ks", "4s", "3d", "5s", "8h", "2c"),
// 		NumOfHoleCards: 7,
// 		NumOfBoard:     0,
// 		HighRanking:    hand.Straight,
// 		LowRanking:     hand.HighCard,
// 	},
// }

// func TestDealingAndEvalutions(t *testing.T) {
// 	t.Parallel()
// 	for _, s := range tests {
// 		g := s.G.get()
// 		hCards, board := dealOut(g, s.Cards)

// 		if len(hCards) != s.NumOfHoleCards {
// 			t.Errorf("%v's number of hole cards = %d; want %d", s.G, len(hCards), s.NumOfHoleCards)
// 		}
// 		if len(board) != s.NumOfBoard {
// 			t.Errorf("%v's number of board cards = %d; want %d", s.G, len(board), s.NumOfBoard)
// 		}

// 		equalRanking := func(h *hand.Hand, r hand.Ranking) bool {
// 			return (h == nil && r == hand.Ranking(0)) || (h.Ranking() == r)
// 		}

// 		highHand := g.FormHighHand(hCards, board)
// 		if !equalRanking(highHand, s.HighRanking) {
// 			t.Errorf("%v's high hand formation = %v; want %v", s.G, highHand.Ranking(), s.HighRanking)
// 		}

// 		lowHand := g.FormLowHand(hCards, board)
// 		if !equalRanking(lowHand, s.LowRanking) {
// 			t.Errorf("%v's low hand formation = %v; want %v", s.G, lowHand.Ranking(), s.LowRanking)
// 		}
// 	}
// }

// func TestBlinds(t *testing.T) {
// 	t.Parallel()

// 	h := Holdem.get()
// 	opts := Config{
// 		Game: Holdem,
// 		Stakes: Stakes{
// 			SmallBet: 5,
// 			BigBet:   10,
// 			Ante:     1,
// 		},
// 		NumOfSeats: 3,
// 		Limit:      NoLimit,
// 	}

// 	// 3 person blinds
// 	holeCards := map[int][]*HoleCard{
// 		0: []*HoleCard{},
// 		1: []*HoleCard{},
// 		2: []*HoleCard{},
// 	}

// 	if h.ForcedBet(holeCards, opts, preflop, 0, 0) != 1 {
// 		t.Fatal("ante")
// 	}
// 	if h.ForcedBet(holeCards, opts, preflop, 1, 1) != 6 {
// 		t.Fatal("small blind and ante")
// 	}
// 	if h.ForcedBet(holeCards, opts, preflop, 2, 2) != 11 {
// 		t.Fatal("big blind and ante")
// 	}

// 	// 2 person blinds
// 	holeCards = map[int][]*HoleCard{
// 		0: []*HoleCard{},
// 		1: []*HoleCard{},
// 	}

// 	if h.ForcedBet(holeCards, opts, preflop, 0, 0) != 6 {
// 		t.Fatal("small blind and ante")
// 	}
// 	if h.ForcedBet(holeCards, opts, preflop, 1, 1) != 11 {
// 		t.Fatal("big blind and ante")
// 	}
// }

// func TestBringIn(t *testing.T) {
// 	t.Parallel()

// 	h := StudHi.get()
// 	opts := Config{
// 		Game: Holdem,
// 		Stakes: Stakes{
// 			SmallBet: 5,
// 			BigBet:   10,
// 			Ante:     1,
// 		},
// 		NumOfSeats: 3,
// 		Limit:      NoLimit,
// 	}

// 	holeCards := map[int][]*HoleCard{
// 		0: []*HoleCard{newHoleCard(hand.AceSpades, Exposed)},
// 		1: []*HoleCard{newHoleCard(hand.TenSpades, Exposed)},
// 		2: []*HoleCard{newHoleCard(hand.TwoSpades, Exposed)},
// 	}

// 	if h.ForcedBet(holeCards, opts, thirdSt, 0, 0) != 1 {
// 		t.Fatal("ante")
// 	}
// 	if h.ForcedBet(holeCards, opts, thirdSt, 1, 1) != 1 {
// 		t.Fatal("ante")
// 	}
// 	if h.ForcedBet(holeCards, opts, thirdSt, 2, 2) != 6 {
// 		t.Fatal("bring in")
// 	}
// }

// func BenchmarkHoldemShowdown(b *testing.B) {
// 	p := pot.New(4)
// 	p.Contribute(0, 100)
// 	p.Contribute(1, 110)
// 	p.Contribute(2, 120)
// 	p.Contribute(3, 130)

// 	for i := 0; i < b.N; i++ {
// 		deck := hand.NewDealer().Deck()
// 		board := deck.PopMulti(5)
// 		holeCards := map[int][]*HoleCard{}
// 		for i := 0; i < 4; i++ {
// 			holeCards[i] = []*HoleCard{
// 				newHoleCard(deck.Pop(), Concealed),
// 				newHoleCard(deck.Pop(), Concealed),
// 			}
// 		}
// 		hCards := cardsFromHoleCardMap(holeCards)
// 		highHands := pot.NewHands(hCards, board, Holdem.get().FormHighHand)
// 		lowHands := pot.NewHands(hCards, board, Holdem.get().FormLowHand)
// 		p.Payout(highHands, lowHands, hand.SortingHigh, 0)
// 	}
// }

// func BenchmarkOmahaHiLoShowdown(b *testing.B) {
// 	p := pot.New(4)
// 	p.Contribute(0, 100)
// 	p.Contribute(1, 110)
// 	p.Contribute(2, 120)
// 	p.Contribute(3, 130)

// 	for i := 0; i < b.N; i++ {
// 		deck := hand.NewDealer().Deck()
// 		board := deck.PopMulti(5)
// 		holeCards := map[int][]*HoleCard{}
// 		for i := 0; i < 4; i++ {
// 			holeCards[i] = []*HoleCard{
// 				newHoleCard(deck.Pop(), Concealed),
// 				newHoleCard(deck.Pop(), Concealed),
// 				newHoleCard(deck.Pop(), Concealed),
// 				newHoleCard(deck.Pop(), Concealed),
// 			}
// 		}
// 		hCards := cardsFromHoleCardMap(holeCards)
// 		highHands := pot.NewHands(hCards, board, OmahaHiLo.get().FormHighHand)
// 		lowHands := pot.NewHands(hCards, board, OmahaHiLo.get().FormLowHand)
// 		p.Payout(highHands, lowHands, hand.SortingHigh, 0)
// 	}
// }

// func dealOut(g game, cards []*hand.Card) (holeCards []*hand.Card, board []*hand.Card) {
// 	deck := jokertest.Dealer(cards).Deck()
// 	hCards := []*HoleCard{}
// 	board = []*hand.Card{}

// 	for i := 0; i < g.NumOfRounds(); i++ {
// 		hCards = append(hCards, g.HoleCards(deck, round(i))...)
// 		board = append(board, g.BoardCards(deck, round(i))...)
// 	}

// 	holeCards = []*hand.Card{}
// 	for _, c := range hCards {
// 		holeCards = append(holeCards, c.Card)
// 	}
// 	return
// }
