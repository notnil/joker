package jokertest

import "github.com/SyntropyDev/joker/hand"

// Cards takes a list of strings that have the format "4s", "Tc",
// "Ah" instead of the hand.Card String() format "4♠", "T♣", "A♥"
// for ease of testing.  If a string is invalid Cards panics,
// otherwise it returns a list of the corresponding cards.
func Cards(list ...string) []*hand.Card {
	cards := []*hand.Card{}
	for _, s := range list {
		cards = append(cards, card(s))
	}
	return cards
}

// Deck returns a hand.Deck that will pop cards in the order of
// the cards given.
func Deck(cards []*hand.Card) hand.Deck {
	// reverse cards
	for i, j := 0, len(cards)-1; i < j; i, j = i+1, j-1 {
		cards[i], cards[j] = cards[j], cards[i]
	}

	return &deck{
		cards: cards,
		input: cards,
	}
}

func card(s string) *hand.Card {
	if len(s) != 2 {
		panic("jokertest: card string must be two characters")
	}

	rank, ok := rankMap[s[:1]]
	if !ok {
		panic("jokertest: rank not found")
	}

	suit, ok := suitMap[s[1:]]
	if !ok {
		panic("jokertest: suit not found")
	}

	for _, c := range hand.Cards() {
		if rank == c.Rank() && suit == c.Suit() {
			return c
		}
	}

	panic("card not found")
}

type deck struct {
	input []*hand.Card
	cards []*hand.Card
}

func (d *deck) Cards() []*hand.Card {
	return append([]*hand.Card(nil), d.cards...)
}

func (d *deck) Discards() []*hand.Card {
	return append([]*hand.Card(nil), d.cards...)
}

func (d *deck) Pop() *hand.Card {
	last := len(d.cards) - 1
	cards, card := d.cards[:last], d.cards[last]
	d.cards = cards
	return card
}

func (d *deck) PopMulti(n int) []*hand.Card {
	cards := []*hand.Card{}
	for i := 0; i < n; i++ {
		cards = append(cards, d.Pop())
	}
	return cards
}

func (d *deck) Reset() {
	d.cards = d.input
}

func (d *deck) Discard(cards ...*hand.Card) {
	panic("not used")
}

func (d *deck) Len() int {
	return len(d.cards)
}

func (d *deck) FromCards(cards, discards []*hand.Card) hand.Deck {
	return Deck(cards)
}

func (d *deck) MarshalJSON() ([]byte, error) {
	return []byte{}, nil
}

func (d *deck) UnmarshalJSON(data []byte) error {
	return nil
}

var (
	rankMap = map[string]hand.Rank{
		"A": hand.Ace,
		"K": hand.King,
		"Q": hand.Queen,
		"J": hand.Jack,
		"T": hand.Ten,
		"9": hand.Nine,
		"8": hand.Eight,
		"7": hand.Seven,
		"6": hand.Six,
		"5": hand.Five,
		"4": hand.Four,
		"3": hand.Three,
		"2": hand.Two,
	}

	suitMap = map[string]hand.Suit{
		"s": hand.Spades,
		"h": hand.Hearts,
		"d": hand.Diamonds,
		"c": hand.Clubs,
	}
)
