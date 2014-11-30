package table

import (
	"testing"

	"github.com/SyntropyDev/joker/hand"
	"github.com/SyntropyDev/joker/jokertest"
)

func TestHoldem(t *testing.T) {
	t.Parallel()

	cards := jokertest.Cards("Qh", "Ks", "4s", "3d", "4s", "8h", "2c", "Ah", "Kh")
	deck := jokertest.Deck(cards)

	h := Holdem.get()
	holeCards := []*HoleCard{}
	for _, r := range holdemRounds() {
		holeCards = append(holeCards, h.HoleCards(deck, r)...)
	}
	if len(holeCards) != 2 {
		t.Fatal("should deal two hole cards")
	}
	boardCards := []*hand.Card{}
	for _, r := range holdemRounds() {
		boardCards = append(boardCards, h.BoardCards(deck, r)...)
	}
	if len(boardCards) != 5 {
		t.Fatal("should deal five board cards")
	}
	plainHoleCards := []*hand.Card{}
	for _, c := range holeCards {
		plainHoleCards = append(plainHoleCards, c.Card)
	}

	_hand := h.FormHighHand(plainHoleCards, boardCards)
	if _hand.Ranking() != hand.Pair {
		t.Fatal(_hand)
	}
}

func TestOmahaHiLo(t *testing.T) {
	t.Parallel()

	cards := jokertest.Cards("Qh", "Ks", "4s", "3d", "4s", "8h", "2c", "Ah", "Kh")
	deck := jokertest.Deck(cards)

	h := OmahaHiLo.get()
	holeCards := []*HoleCard{}
	for _, r := range holdemRounds() {
		holeCards = append(holeCards, h.HoleCards(deck, r)...)
	}
	if len(holeCards) != 4 {
		t.Fatal("should deal four hole cards")
	}
	boardCards := []*hand.Card{}
	for _, r := range holdemRounds() {
		boardCards = append(boardCards, h.BoardCards(deck, r)...)
	}
	if len(boardCards) != 5 {
		t.Fatal("should deal five board cards")
	}
	plainHoleCards := []*hand.Card{}
	for _, c := range holeCards {
		plainHoleCards = append(plainHoleCards, c.Card)
	}

	highHand := h.FormHighHand(plainHoleCards, boardCards)
	if highHand.Ranking() != hand.TwoPair {
		t.Fatal("should find a two pair")
	}

	lowHand := h.FormLowHand(plainHoleCards, boardCards)
	if lowHand == nil {
		t.Fatal("should find low")
	}
}

func TestStudHiLo(t *testing.T) {
	t.Parallel()

	cards := jokertest.Cards("Qh", "Ks", "4s", "3d", "4s", "8h", "2c", "Ah", "Kh")
	deck := jokertest.Deck(cards)

	h := StudHiLo.get()
	holeCards := []*HoleCard{}
	for _, r := range studRounds() {
		holeCards = append(holeCards, h.HoleCards(deck, r)...)
	}
	if len(holeCards) != 7 {
		t.Fatal("should deal seven hole cards")
	}
	boardCards := []*hand.Card{}
	for _, r := range studRounds() {
		boardCards = append(boardCards, h.BoardCards(deck, r)...)
	}
	if len(boardCards) != 0 {
		t.Fatal("should deal zero board cards")
	}
	plainHoleCards := []*hand.Card{}
	for _, c := range holeCards {
		plainHoleCards = append(plainHoleCards, c.Card)
	}

	highHand := h.FormHighHand(plainHoleCards, boardCards)

	if highHand.Ranking() != hand.Pair {
		t.Fatal("should find a pair")
	}

	lowHand := h.FormLowHand(plainHoleCards, boardCards)
	if lowHand != nil {
		t.Fatal("should not find low")
	}
}

func TestBlinds(t *testing.T) {
	t.Parallel()

	h := Holdem.get()
	opts := Options{
		Game: Holdem,
		Stakes: Stakes{
			SmallBet: 5,
			BigBet:   10,
			Ante:     1,
		},
		NumOfSeats: 3,
		Limit:      NoLimit,
	}

	// 3 person blinds
	holeCards := map[int][]*HoleCard{
		0: []*HoleCard{},
		1: []*HoleCard{},
		2: []*HoleCard{},
	}

	if h.ForcedBet(holeCards, opts, preflop, 0, 0) != 1 {
		t.Fatal("ante")
	}
	if h.ForcedBet(holeCards, opts, preflop, 1, 1) != 6 {
		t.Fatal("small blind and ante")
	}
	if h.ForcedBet(holeCards, opts, preflop, 2, 2) != 11 {
		t.Fatal("big blind and ante")
	}

	// 2 person blinds

	holeCards = map[int][]*HoleCard{
		0: []*HoleCard{},
		1: []*HoleCard{},
	}

	if h.ForcedBet(holeCards, opts, preflop, 0, 0) != 6 {
		t.Fatal("small blind and ante")
	}
	if h.ForcedBet(holeCards, opts, preflop, 1, 1) != 11 {
		t.Fatal("big blind and ante")
	}
}

func TestBringIn(t *testing.T) {
	t.Parallel()

	h := StudHi.get()
	opts := Options{
		Game: Holdem,
		Stakes: Stakes{
			SmallBet: 5,
			BigBet:   10,
			Ante:     1,
		},
		NumOfSeats: 3,
		Limit:      NoLimit,
	}

	holeCards := map[int][]*HoleCard{
		0: []*HoleCard{newHoleCard(hand.AceSpades, Exposed)},
		1: []*HoleCard{newHoleCard(hand.TenSpades, Exposed)},
		2: []*HoleCard{newHoleCard(hand.TwoSpades, Exposed)},
	}

	if h.ForcedBet(holeCards, opts, thirdSt, 0, 0) != 1 {
		t.Fatal("ante")
	}
	if h.ForcedBet(holeCards, opts, thirdSt, 1, 1) != 1 {
		t.Fatal("ante")
	}
	if h.ForcedBet(holeCards, opts, thirdSt, 2, 2) != 6 {
		t.Fatal("bring in")
	}
}

func BenchmarkOmahaHiLoShowdown(b *testing.B) {
	pot := newPot(4)
	pot.contribute(0, 100)
	pot.contribute(1, 110)
	pot.contribute(2, 120)
	pot.contribute(3, 130)

	for i := 0; i < b.N; i++ {
		deck := hand.NewDeck()
		board := deck.PopMulti(5)
		holeCards := map[int][]*HoleCard{}
		for i := 0; i < 4; i++ {
			holeCards[i] = []*HoleCard{
				newHoleCard(deck.Pop(), Concealed),
				newHoleCard(deck.Pop(), Concealed),
				newHoleCard(deck.Pop(), Concealed),
				newHoleCard(deck.Pop(), Concealed),
			}
		}
		highHands := newHands(holeCards, board, OmahaHiLo.get().FormHighHand)
		lowHands := newHands(holeCards, board, OmahaHiLo.get().FormLowHand)
		pot.payout(highHands, lowHands, hand.SortingHigh, true, 0)
	}
}
