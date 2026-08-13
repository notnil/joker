package hand_test

import (
	"testing"

	"github.com/notnil/joker/pkg/hand"
	. "github.com/notnil/joker/pkg/jokertest"
)

func TestRankingString(t *testing.T) {
	cases := []struct {
		r    hand.Ranking
		want string
	}{
		{hand.HighCard, "HighCard"},
		{hand.Pair, "Pair"},
		{hand.TwoPair, "TwoPair"},
		{hand.ThreeOfAKind, "ThreeOfAKind"},
		{hand.Straight, "Straight"},
		{hand.Flush, "Flush"},
		{hand.FullHouse, "FullHouse"},
		{hand.FourOfAKind, "FourOfAKind"},
		{hand.StraightFlush, "StraightFlush"},
		{hand.RoyalFlush, "RoyalFlush"},
		{hand.Ranking(0), "Ranking(0)"},
		{hand.Ranking(11), "Ranking(11)"},
	}
	for _, tc := range cases {
		if got := tc.r.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", tc.r, got, tc.want)
		}
	}
}

func TestEmptyHand(t *testing.T) {
	h := hand.New(nil)
	if h.Ranking() != hand.HighCard {
		t.Fatalf("empty ranking = %v, want HighCard", h.Ranking())
	}
	if len(h.Cards()) != 0 {
		t.Fatalf("empty cards = %v, want none", h.Cards())
	}
	if h.Description() != "no cards" {
		t.Fatalf("description = %q", h.Description())
	}
	other := hand.New(Cards("As"))
	if h.CompareTo(other) >= 0 {
		t.Fatal("empty hand should lose to a card")
	}
}

func TestDeckTextRoundTrip(t *testing.T) {
	d := &hand.Deck{Cards: Cards("As", "Kh", "2c")}
	text, err := d.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	got := &hand.Deck{}
	if err := got.UnmarshalText(text); err != nil {
		t.Fatal(err)
	}
	if len(got.Cards) != 3 {
		t.Fatalf("len = %d", len(got.Cards))
	}
	for i, c := range d.Cards {
		if got.Cards[i] != c {
			t.Fatalf("card %d = %v, want %v", i, got.Cards[i], c)
		}
	}
}

func TestFamousHands(t *testing.T) {
	royal := hand.New(Cards("As", "Ks", "Qs", "Js", "Ts", "2h", "3d"))
	if royal.Ranking() != hand.RoyalFlush {
		t.Fatalf("royal = %v", royal.Ranking())
	}
	quads := hand.New(Cards("Ah", "Ad", "Ac", "As", "Kh", "Kd", "2c"))
	if quads.Ranking() != hand.FourOfAKind {
		t.Fatalf("quads = %v", quads.Ranking())
	}
	if royal.CompareTo(quads) <= 0 {
		t.Fatal("royal should beat quads")
	}
	wheel := hand.New(Cards("As", "2h", "3d", "4c", "5s"))
	broadway := hand.New(Cards("As", "Kh", "Qd", "Jc", "Ts"))
	if wheel.Ranking() != hand.Straight || broadway.Ranking() != hand.Straight {
		t.Fatal("expected two straights")
	}
	if broadway.CompareTo(wheel) <= 0 {
		t.Fatal("broadway should beat wheel")
	}
	chopA := hand.New(Cards("As", "Ks", "Qs", "Js", "9d", "2h", "3c"))
	chopB := hand.New(Cards("Ah", "Kh", "Qh", "Jh", "9c", "4d", "5s"))
	if chopA.CompareTo(chopB) != 0 {
		t.Fatal("same high-card kickers should chop")
	}
}
