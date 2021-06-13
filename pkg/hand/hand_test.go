package hand_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/notnil/joker/hand"
	. "github.com/notnil/joker/jokertest"
)

type testPair struct {
	cards       []hand.Card
	arrangement []hand.Card
	ranking     hand.Ranking
	description string
}

var tests = []testPair{
	{
		Cards("Ks", "Qs", "Js", "As", "9d"),
		Cards("As", "Ks", "Qs", "Js", "9d"),
		hand.HighCard,
		"high card ace high",
	},
	{
		Cards("Ks", "Qh", "Qs", "Js", "9d"),
		Cards("Qh", "Qs", "Ks", "Js", "9d"),
		hand.Pair,
		"pair of queens",
	},
	{
		Cards("2s", "Qh", "Qs", "Js", "2d"),
		Cards("Qh", "Qs", "2s", "2d", "Js"),
		hand.TwoPair,
		"two pair queens and twos",
	},
	{
		Cards("6s", "Qh", "Ks", "6h", "6d"),
		Cards("6s", "6h", "6d", "Ks", "Qh"),
		hand.ThreeOfAKind,
		"three of a kind sixes",
	},
	{
		Cards("Ks", "Qs", "Js", "As", "Td"),
		Cards("As", "Ks", "Qs", "Js", "Td"),
		hand.Straight,
		"straight ace high",
	},
	{
		Cards("2s", "3s", "4s", "As", "5d"),
		Cards("5d", "4s", "3s", "2s", "As"),
		hand.Straight,
		"straight five high",
	},
	{
		Cards("7s", "4s", "5s", "3s", "2s"),
		Cards("7s", "5s", "4s", "3s", "2s"),
		hand.Flush,
		"flush seven high",
	},
	{
		Cards("7s", "7d", "3s", "3d", "7h"),
		Cards("7s", "7d", "7h", "3s", "3d"),
		hand.FullHouse,
		"full house sevens full of threes",
	},
	{
		Cards("7s", "7d", "3s", "7c", "7h"),
		Cards("7s", "7d", "7c", "7h", "3s"),
		hand.FourOfAKind,
		"four of a kind sevens",
	},
	{
		Cards("Ks", "Qs", "Js", "Ts", "9s"),
		Cards("Ks", "Qs", "Js", "Ts", "9s"),
		hand.StraightFlush,
		"straight flush king high",
	},
	{
		Cards("As", "5s", "4s", "3s", "2s"),
		Cards("5s", "4s", "3s", "2s", "As"),
		hand.StraightFlush,
		"straight flush five high",
	},
	{
		Cards("As", "Ks", "Qs", "Js", "Ts"),
		Cards("As", "Ks", "Qs", "Js", "Ts"),
		hand.RoyalFlush,
		"royal flush",
	},
	{
		Cards("As", "Ks", "Qs", "2s", "2c", "2h", "2d"),
		Cards("2s", "2c", "2h", "2d", "As"),
		hand.FourOfAKind,
		"four of a kind twos",
	},
}

func TestHands(t *testing.T) {
	for _, test := range tests {
		h := hand.New(test.cards)
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
	cards1 []hand.Card
	cards2 []hand.Card
	e      equality
}

var equalityTests = []testEquality{
	{
		Cards("As", "5s", "4s", "3s", "2s"),
		Cards("Ks", "Kc", "Kh", "Jd", "Js"),
		greaterThan,
	},
	{
		Cards("Ts", "9h", "8d", "7c", "6s", "2h", "3s"),
		Cards("Ts", "9h", "8d", "7c", "6s", "Ah", "Ks"),
		equalTo,
	},
}

func TestCompareHands(t *testing.T) {
	for _, test := range equalityTests {
		h1 := hand.New(test.cards1)
		h2 := hand.New(test.cards2)
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
	cards       []hand.Card
	arrangement []hand.Card
	options     []func(*hand.Config)
	ranking     hand.Ranking
	description string
}

var optTests = []testOptionsPairs{
	{
		Cards("Ks", "Qs", "Js", "As", "9s"),
		Cards("As", "Ks", "Qs", "Js", "9s"),
		[]func(*hand.Config){hand.Low},
		hand.Flush,
		"flush ace high",
	},
	{
		Cards("7h", "6h", "5s", "4s", "2s", "3s"),
		Cards("6h", "5s", "4s", "3s", "2s"),
		[]func(*hand.Config){hand.AceToFiveLow},
		hand.HighCard,
		"high card six high",
	},
	{
		Cards("Ah", "6h", "5s", "4s", "2s", "Ks"),
		Cards("6h", "5s", "4s", "2s", "Ah"),
		[]func(*hand.Config){hand.AceToFiveLow},
		hand.HighCard,
		"high card six high",
	},
}

func TestHandsWithOptions(t *testing.T) {
	for _, test := range optTests {
		h := hand.New(test.cards, test.options...)
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
	cards := []hand.Card{hand.AceSpades}
	h := hand.New(cards)
	if h.Ranking() != hand.HighCard {
		t.Fatal("blank card error")
	}

	cards = []hand.Card{hand.FiveSpades, hand.FiveClubs}
	h = hand.New(cards)
	if h.Ranking() != hand.Pair {
		t.Fatal("blank card error")
	}
}

func TestDeck(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	deck := hand.NewDealer(r).Deck()
	if deck.Pop() == deck.Pop() {
		t.Fatal("Two Pop() calls should never return the same result")
	}
	l := len(deck.Cards)
	if l != 50 {
		t.Fatalf("After Pop() deck len = %d; want %d", l, 50)
	}
}

func TestHandJSON(t *testing.T) {
	jsonStr := `{"ranking":10,"cards":["A♠","K♠","Q♠","J♠","T♠"],"description":"royal flush","config":{"sorting":1,"ignoreStraights":false,"ignoreFlushes":false,"aceIsLow":false}}`
	h := &hand.Hand{}
	if err := json.Unmarshal([]byte(jsonStr), h); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != jsonStr {
		t.Fatalf("expected json %s but got %s", jsonStr, string(b))
	}
}

func BenchmarkHandCreation(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	cards := hand.NewDealer(r).Deck().PopMulti(7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hand.New(cards)
	}
}
