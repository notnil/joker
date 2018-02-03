package hand

import (
	"errors"
	"strings"
)

// A Rank represents the rank of a card.
type Rank int

const (
	Two Rank = iota
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

const (
	ranksStr = "23456789TJQKA"
)

var (
	singularNames = []string{"two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "jack", "queen", "king", "ace"}
	pluralNames   = []string{"twos", "threes", "fours", "fives", "sixes", "sevens", "eights", "nines", "tens", "jacks", "queens", "kings", "aces"}
)

// String returns a string in the format "2"
func (r Rank) String() string {
	return ranksStr[r : r+1]
}

// singularName returns the name of the rank in singular form such as "two" for Two.
func (r Rank) singularName() string {
	return singularNames[r]
}

// pluralName returns the name of the rank in plural form such as "twos" for Two.
func (r Rank) pluralName() string {
	return pluralNames[r]
}

func (r Rank) aceLowIndexOf() int {
	for i, rank := range allAceLowRanks() {
		if r == rank {
			return i
		}
	}
	return -1
}

type byAceHighRank []Rank

func (a byAceHighRank) Len() int { return len(a) }

func (a byAceHighRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byAceHighRank) Less(i, j int) bool {
	return a[i] < a[j]
}

type byAceLowRank []Rank

func (a byAceLowRank) Len() int { return len(a) }

func (a byAceLowRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byAceLowRank) Less(i, j int) bool {
	iRank, jRank := a[i], a[j]
	iIndex, jIndex := iRank.aceLowIndexOf(), jRank.aceLowIndexOf()
	return iIndex < jIndex
}

// A Suit represents the suit of a card.
type Suit int

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

var (
	suitsStr = []string{"♠", "♥", "♦", "♣"}
	suitsMap = map[string]Suit{
		"♠": Spades,
		"♥": Hearts,
		"♦": Diamonds,
		"♣": Clubs,
	}
)

// String returns a string in the format "♠"
func (s Suit) String() string {
	return suitsStr[s]
}

type Card int

const (
	TwoSpades Card = iota
	ThreeSpades
	FourSpades
	FiveSpades
	SixSpades
	SevenSpades
	EightSpades
	NineSpades
	TenSpades
	JackSpades
	QueenSpades
	KingSpades
	AceSpades

	TwoHearts
	ThreeHearts
	FourHearts
	FiveHearts
	SixHearts
	SevenHearts
	EightHearts
	NineHearts
	TenHearts
	JackHearts
	QueenHearts
	KingHearts
	AceHearts

	TwoDiamonds
	ThreeDiamonds
	FourDiamonds
	FiveDiamonds
	SixDiamonds
	SevenDiamonds
	EightDiamonds
	NineDiamonds
	TenDiamonds
	JackDiamonds
	QueenDiamonds
	KingDiamonds
	AceDiamonds

	TwoClubs
	ThreeClubs
	FourClubs
	FiveClubs
	SixClubs
	SevenClubs
	EightClubs
	NineClubs
	TenClubs
	JackClubs
	QueenClubs
	KingClubs
	AceClubs
)

func getCard(r Rank, s Suit) Card {
	return Card(int(r) + (13 * int(s)))
}

// Rank returns the rank of the card.
func (c Card) Rank() Rank {
	return Rank(c % 13)
}

// Suit returns the suit of the card.
func (c Card) Suit() Suit {
	return Suit(c / 13)
}

// String returns a string in the format "4♠"
func (c Card) String() string {
	return c.Rank().String() + c.Suit().String()
}

// MarshalText implements the encoding.TextMarshaler interface.
// The text format is "4♠".
func (c Card) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The card is expected to be in the format "4♠".
func (c *Card) UnmarshalText(text []byte) error {
	s := string(text)
	if len(s) <= 1 {
		return errors.New("hand: invalid card text " + s)
	}
	rank := strings.Index(ranksStr, s[0:1])
	suit, ok := suitsMap[s[1:]]
	if !ok {
		return errors.New("hand: invalid suit " + s[1:])
	}
	if rank == -1 {
		return errors.New("hand: invalid rank " + s[0:1])
	}
	card := getCard(Rank(rank), suit)
	c = &card
	return nil
}

// Cards returns all 52 unshuffled cards
func Cards() []Card {
	return []Card{
		AceSpades, KingSpades, QueenSpades, JackSpades, TenSpades,
		NineSpades, EightSpades, SevenSpades, SixSpades, FiveSpades,
		FourSpades, ThreeSpades, TwoSpades,

		AceHearts, KingHearts, QueenHearts, JackHearts, TenHearts,
		NineHearts, EightHearts, SevenHearts, SixHearts, FiveHearts,
		FourHearts, ThreeHearts, TwoHearts,

		AceDiamonds, KingDiamonds, QueenDiamonds, JackDiamonds, TenDiamonds,
		NineDiamonds, EightDiamonds, SevenDiamonds, SixDiamonds, FiveDiamonds,
		FourDiamonds, ThreeDiamonds, TwoDiamonds,

		AceClubs, KingClubs, QueenClubs, JackClubs, TenClubs,
		NineClubs, EightClubs, SevenClubs, SixClubs, FiveClubs,
		FourClubs, ThreeClubs, TwoClubs,
	}
}

type byAceHigh []Card

func (a byAceHigh) Len() int { return len(a) }

func (a byAceHigh) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byAceHigh) Less(i, j int) bool {
	return a[i].Rank() < a[j].Rank()
}

type byAceLow []Card

func (a byAceLow) Len() int { return len(a) }

func (a byAceLow) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byAceLow) Less(i, j int) bool {
	iCard, jCard := a[i], a[j]
	iIndex, jIndex := iCard.Rank().aceLowIndexOf(), jCard.Rank().aceLowIndexOf()
	return iIndex < jIndex
}

func allRanks() []Rank {
	return []Rank{Two, Three, Four, Five, Six, Seven, Eight,
		Nine, Ten, Jack, Queen, King, Ace}
}

func allAceLowRanks() []Rank {
	return []Rank{Ace, Two, Three, Four, Five, Six, Seven, Eight,
		Nine, Ten, Jack, Queen, King}
}

func allSuits() []Suit {
	return []Suit{Spades, Hearts, Diamonds, Clubs}
}
