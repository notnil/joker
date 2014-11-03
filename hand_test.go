package joker_test

import (
	"testing"

	. "github.com/SyntropyDev/joker"
	"github.com/SyntropyDev/jokertest"
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
		h := NewHand(test.cards)
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
		[]*Card{AceSpades, FiveSpades, FourSpades, ThreeSpades, TwoSpades},
		[]*Card{KingSpades, KingClubs, KingHearts, JackDiamonds, JackSpades},
		greaterThan,
	},
	{
		[]*Card{TenSpades, NineHearts, EightDiamonds, SevenClubs, SixSpades, TwoHearts, ThreeSpades},
		[]*Card{TenSpades, NineHearts, EightDiamonds, SevenClubs, SixSpades, AceHearts, KingSpades},
		equalTo,
	},
}

func TestCompareHands(t *testing.T) {
	for _, test := range equalityTests {
		h1 := NewHand(test.cards1)
		h2 := NewHand(test.cards2)
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
	options     Options
	ranking     Ranking
	description string
}

var optTests = []testOptionsPairs{
	{
		jokertest.Cards("Ks", "Qs", "Js", "As", "9s"),
		jokertest.Cards("As", "Ks", "Qs", "Js", "9s"),
		Options{
			Sorting:         Low,
			IgnoreStraights: true,
			IgnoreFlushes:   true,
			AceIsLow:        false,
		},
		HighCard,
		"high card ace high",
	},
	{
		jokertest.Cards("7h", "6h", "5s", "4s", "2s", "3s"),
		jokertest.Cards("6h", "5s", "4s", "3s", "2s"),
		Options{
			Sorting:         Low,
			IgnoreStraights: true,
			IgnoreFlushes:   true,
			AceIsLow:        true,
		},
		HighCard,
		"high card six high",
	},
	{
		jokertest.Cards("Ah", "6h", "5s", "4s", "2s", "Ks"),
		jokertest.Cards("6h", "5s", "4s", "2s", "Ah"),
		Options{
			Sorting:         Low,
			IgnoreStraights: true,
			IgnoreFlushes:   true,
			AceIsLow:        true,
		},
		HighCard,
		"high card six high",
	},
}

func TestHandsWithOptions(t *testing.T) {
	for _, test := range optTests {
		h := NewHandWithOptions(test.cards, test.options)
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
	hand := NewHand(cards)
	if hand.Ranking() != HighCard {
		t.Fatal("blank card error")
	}

	cards = []*Card{FiveSpades, FiveClubs}
	hand = NewHand(cards)
	if hand.Ranking() != Pair {
		t.Fatal("blank card error")
	}
}

func BenchmarkHandCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cards := NewDeck().PopMulti(7)
		NewHand(cards)
	}
}
