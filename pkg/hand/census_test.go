package hand_test

import (
	"testing"

	"github.com/notnil/joker/pkg/hand"
)

// Cactus Kev / Wikipedia 5-card category counts. Straight flushes include
// the 4 royal flushes (40 total).
func TestFiveCardCensus(t *testing.T) {
	deck := all52()
	counts := map[hand.Ranking]int{}
	n := 0
	forEachFive(deck, func(cards [5]hand.Card) {
		h := hand.New(cards[:])
		counts[h.Ranking()]++
		n++
	})
	if n != 2598960 {
		t.Fatalf("enumerated %d hands, want 2598960", n)
	}
	want := map[hand.Ranking]int{
		hand.HighCard:      1302540,
		hand.Pair:          1098240,
		hand.TwoPair:       123552,
		hand.ThreeOfAKind:  54912,
		hand.Straight:      10200,
		hand.Flush:         5108,
		hand.FullHouse:     3744,
		hand.FourOfAKind:   624,
		hand.StraightFlush: 36,
		hand.RoyalFlush:    4,
	}
	for r, w := range want {
		if counts[r] != w {
			t.Errorf("%s: got %d, want %d", r, counts[r], w)
		}
	}
	if counts[hand.StraightFlush]+counts[hand.RoyalFlush] != 40 {
		t.Errorf("straight flushes including royals = %d, want 40",
			counts[hand.StraightFlush]+counts[hand.RoyalFlush])
	}
}

func all52() []hand.Card {
	cards := make([]hand.Card, 0, 52)
	for s := hand.Spades; s <= hand.Clubs; s++ {
		for r := hand.Two; r <= hand.Ace; r++ {
			for _, c := range hand.Cards() {
				if c.Rank() == r && c.Suit() == s {
					cards = append(cards, c)
					break
				}
			}
		}
	}
	if len(cards) != 52 {
		panic("expected 52 cards")
	}
	return cards
}

func forEachFive(deck []hand.Card, fn func([5]hand.Card)) {
	n := len(deck)
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			for c := b + 1; c < n; c++ {
				for d := c + 1; d < n; d++ {
					for e := d + 1; e < n; e++ {
						fn([5]hand.Card{deck[a], deck[b], deck[c], deck[d], deck[e]})
					}
				}
			}
		}
	}
}
