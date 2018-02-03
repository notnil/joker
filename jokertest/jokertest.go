package jokertest

import "github.com/notnil/joker/hand"

// Cards takes a list of strings that have the format "4s", "Tc",
// "Ah" instead of the hand.Card String() format "4♠", "T♣", "A♥"
// for ease of testing.  If a string is invalid Cards panics,
// otherwise it returns a list of the corresponding cards.
func Cards(list ...string) []hand.Card {
	cards := []hand.Card{}
	for _, s := range list {
		cards = append(cards, card(s))
	}
	return cards
}

// Dealer returns a hand.Dealer that generates decks that will pop
// cards in the order of the cards given.
func Dealer(cards []hand.Card) hand.Dealer {
	return &deck{cards: cards}
}

type deck struct {
	cards []hand.Card
}

func (d deck) Deck() *hand.Deck {
	// copy cards
	cards := make([]hand.Card, len(d.cards))
	copy(cards, d.cards)

	// reverse cards
	for i, j := 0, len(cards)-1; i < j; i, j = i+1, j-1 {
		cards[i], cards[j] = cards[j], cards[i]
	}
	return &hand.Deck{Cards: cards}
}

func card(s string) hand.Card {
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
	panic("jokertest: card not found")
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
