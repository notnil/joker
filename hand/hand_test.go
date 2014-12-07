package hand_test

import (
	"encoding/json"
	"testing"

	. "github.com/SyntropyDev/joker/hand"
	"github.com/SyntropyDev/joker/jokertest"
)

type testPair struct {
	cards       []*Card
	arrangement []*Card
	ranking     Ranking
	description string
}

var tests = []testPair{
	{
		jokertest.Cards("Ks", "Qs", "Js", "As", "9d"),
		jokertest.Cards("As", "Ks", "Qs", "Js", "9d"),
		HighCard,
		"high card ace high",
	},
	{
		jokertest.Cards("Ks", "Qh", "Qs", "Js", "9d"),
		jokertest.Cards("Qh", "Qs", "Ks", "Js", "9d"),
		Pair,
		"pair of queens",
	},
	{
		jokertest.Cards("2s", "Qh", "Qs", "Js", "2d"),
		jokertest.Cards("Qh", "Qs", "2s", "2d", "Js"),
		TwoPair,
		"two pair queens and twos",
	},
	{
		jokertest.Cards("6s", "Qh", "Ks", "6h", "6d"),
		jokertest.Cards("6s", "6h", "6d", "Ks", "Qh"),
		ThreeOfAKind,
		"three of a kind sixes",
	},
	{
		jokertest.Cards("Ks", "Qs", "Js", "As", "Td"),
		jokertest.Cards("As", "Ks", "Qs", "Js", "Td"),
		Straight,
		"straight ace high",
	},
	{
		jokertest.Cards("2s", "3s", "4s", "As", "5d"),
		jokertest.Cards("5d", "4s", "3s", "2s", "As"),
		Straight,
		"straight five high",
	},
	{
		jokertest.Cards("7s", "4s", "5s", "3s", "2s"),
		jokertest.Cards("7s", "5s", "4s", "3s", "2s"),
		Flush,
		"flush seven high",
	},
	{
		jokertest.Cards("7s", "7d", "3s", "3d", "7h"),
		jokertest.Cards("7s", "7d", "7h", "3s", "3d"),
		FullHouse,
		"full house sevens full of threes",
	},
	{
		jokertest.Cards("7s", "7d", "3s", "7c", "7h"),
		jokertest.Cards("7s", "7d", "7c", "7h", "3s"),
		FourOfAKind,
		"four of a kind sevens",
	},
	{
		jokertest.Cards("Ks", "Qs", "Js", "Ts", "9s"),
		jokertest.Cards("Ks", "Qs", "Js", "Ts", "9s"),
		StraightFlush,
		"straight flush king high",
	},
	{
		jokertest.Cards("As", "5s", "4s", "3s", "2s"),
		jokertest.Cards("5s", "4s", "3s", "2s", "As"),
		StraightFlush,
		"straight flush five high",
	},
	{
		jokertest.Cards("As", "Ks", "Qs", "Js", "Ts"),
		jokertest.Cards("As", "Ks", "Qs", "Js", "Ts"),
		RoyalFlush,
		"royal flush",
	},
	{
		jokertest.Cards("As", "Ks", "Qs", "2s", "2c", "2h", "2d"),
		jokertest.Cards("2s", "2c", "2h", "2d", "As"),
		FourOfAKind,
		"four of a kind twos",
	},
}

func TestHands(t *testing.T) {
	for _, test := range tests {
		h := New(test.cards)
		if h.Ranking() != test.ranking {
			t.Fatalf("expected %v got %v", test.ranking, h.Ranking())
		}
		for i := 0; i < 5; i++ {
			actual, expected := h.Cards()[i], test.arrangement[i]
			if actual.Rank() != expected.Rank() || actual.Suit() != expected.Suit() {
				t.Fatalf("expected %v got %v", expected, actual)
			}
		}
		if test.description != h.Description() {
			t.Fatalf("expected \"%v\" got \"%v\"", test.description, h.Description())
		}
	}
}

type equality int

const (
	greaterThan equality = iota
	lessThan
	equalTo
)

type testEquality struct {
	cards1 []*Card
	cards2 []*Card
	e      equality
}

var equalityTests = []testEquality{
	{
		jokertest.Cards("As", "5s", "4s", "3s", "2s"),
		jokertest.Cards("Ks", "Kc", "Kh", "Jd", "Js"),
		greaterThan,
	},
	{
		jokertest.Cards("Ts", "9h", "8d", "7c", "6s", "2h", "3s"),
		jokertest.Cards("Ts", "9h", "8d", "7c", "6s", "Ah", "Ks"),
		equalTo,
	},
}

func TestCompareHands(t *testing.T) {
	for _, test := range equalityTests {
		h1 := New(test.cards1)
		h2 := New(test.cards2)
		compareTo := h1.CompareTo(h2)

		switch test.e {
		case greaterThan:
			if compareTo <= 0 {
				t.Errorf("expected %v to be greater than %v", h1, h2)
			}
		case lessThan:
			if compareTo >= 0 {
				t.Errorf("expected %v to be less than %v", h1, h2)
			}
		case equalTo:
			if compareTo != 0 {
				t.Errorf("expected %v to be equal to %v", h1, h2)
			}
		}
	}
}

type testOptionsPairs struct {
	cards       []*Card
	arrangement []*Card
	options     []func(*Config)
	ranking     Ranking
	description string
}

var optTests = []testOptionsPairs{
	{
		jokertest.Cards("Ks", "Qs", "Js", "As", "9s"),
		jokertest.Cards("As", "Ks", "Qs", "Js", "9s"),
		[]func(*Config){Low},
		Flush,
		"flush ace high",
	},
	{
		jokertest.Cards("7h", "6h", "5s", "4s", "2s", "3s"),
		jokertest.Cards("6h", "5s", "4s", "3s", "2s"),
		[]func(*Config){AceToFiveLow},
		HighCard,
		"high card six high",
	},
	{
		jokertest.Cards("Ah", "6h", "5s", "4s", "2s", "Ks"),
		jokertest.Cards("6h", "5s", "4s", "2s", "Ah"),
		[]func(*Config){AceToFiveLow},
		HighCard,
		"high card six high",
	},
}

func TestHandsWithOptions(t *testing.T) {
	for _, test := range optTests {
		h := New(test.cards, test.options...)
		if h.Ranking() != test.ranking {
			t.Fatalf("expected %v got %v", test.ranking, h.Ranking())
		}
		for i := 0; i < 5; i++ {
			actual, expected := h.Cards()[i], test.arrangement[i]
			if actual.Rank() != expected.Rank() || actual.Suit() != expected.Suit() {
				t.Fatalf("expected %v got %v", expected, actual)
			}
		}
		if test.description != h.Description() {
			t.Fatalf("expected \"%v\" got \"%v\"", test.description, h.Description())
		}
	}
}

func TestBlanks(t *testing.T) {
	cards := []*Card{AceSpades}
	hand := New(cards)
	if hand.Ranking() != HighCard {
		t.Fatal("blank card error")
	}

	cards = []*Card{FiveSpades, FiveClubs}
	hand = New(cards)
	if hand.Ranking() != Pair {
		t.Fatal("blank card error")
	}
}

func TestDeck(t *testing.T) {
	deck := NewDealer().Deck()
	if deck.Pop() == deck.Pop() {
		t.Fatal("Two Pop() calls should never return the same result")
	}
	l := len(deck.Cards)
	if l != 50 {
		t.Fatalf("After Pop() deck len = %d; want %d", l, 50)
	}
}

func TestCardJSON(t *testing.T) {
	card := AceSpades

	// to json
	b, err := json.Marshal(card)
	if err != nil {
		t.Fatal(err)
	}

	// and back
	cardCopy := KingHearts
	if err := json.Unmarshal(b, cardCopy); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkHandCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cards := NewDealer().Deck().PopMulti(7)
		New(cards)
	}
}
