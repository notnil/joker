package jokertest

import (
	"strings"

	"github.com/notnil/joker/hand"
)

const (
	deck1Str = "J♦ 7♥ 2♥ 3♣ 5♥ 5♦ 2♦ 3♦ Q♦ 9♠ A♣ 9♣ T♠ 7♦ J♥ 4♦ A♦ J♣ K♣ 9♥ T♦ 2♠ 6♣ 2♣ 6♦ 7♣ 8♣ K♠ 8♠ 6♠ 5♠ 6♥ Q♠ 5♣ Q♣ Q♥ 4♣ 3♠ A♠ 8♦ K♦ 9♦ 4♥ K♥ 8♥ T♥ 3♥ A♥ T♣ J♠ 7♠ 4♠"
	deck2Str = "A♥ 4♥ 3♣ 2♥ 9♥ 9♠ 3♦ 8♥ 2♦ 6♣ 2♣ T♥ K♦ 9♣ 7♣ 7♠ 6♦ J♣ 8♠ J♥ Q♣ 6♥ T♠ A♦ 8♦ 8♣ J♦ 5♦ Q♦ A♣ 2♠ T♦ K♣ A♠ Q♠ 6♠ 4♦ 5♠ Q♥ 3♠ K♥ 7♦ 5♥ 4♣ 3♥ 9♦ 7♥ J♠ K♠ 4♠ 5♣ T♣"
	deck3Str = "Q♠ 5♦ 5♣ 4♦ Q♣ 4♥ 7♥ Q♥ T♦ 2♦ 4♠ T♣ J♣ A♣ 8♥ 3♠ 7♣ 9♥ 8♣ 9♣ 6♦ 6♣ 8♦ A♦ K♥ J♦ 7♠ 2♥ 7♦ 3♥ A♥ 9♠ K♣ 2♣ 8♠ 5♠ 6♥ T♠ T♥ A♠ 3♦ 9♦ 6♠ K♦ J♥ K♠ 4♣ 3♣ J♠ Q♦ 5♥ 2♠"
	deck4Str = "6♠ J♠ J♣ 8♣ Q♥ A♣ T♦ T♠ Q♦ 5♠ Q♠ 9♠ 4♦ 7♦ 3♥ 4♣ 8♥ A♥ 6♣ 7♣ T♣ 7♠ K♣ 3♠ 4♥ K♥ 9♥ 5♥ 6♦ 3♣ 3♦ 7♥ 2♣ 6♥ T♥ 9♣ J♦ 9♦ 2♦ K♦ 8♠ K♠ 4♠ J♥ Q♣ 2♠ 2♥ 5♦ 8♦ A♦ 5♣ A♠"
	deck5Str = "5♥ 6♥ 6♣ 3♠ T♣ Q♣ 5♦ A♦ 5♣ J♦ 9♦ 9♣ A♣ 8♠ K♥ 8♦ 7♣ K♣ T♦ 2♥ Q♦ 5♠ Q♥ K♠ 8♣ 4♥ 3♣ K♦ 2♣ T♠ T♥ 8♥ 4♣ Q♠ 4♦ A♥ 3♦ 6♠ 9♠ A♠ 2♠ 7♠ 2♦ 9♥ 4♠ 6♦ 3♥ J♣ 7♦ J♥ 7♥ J♠"
)

// Deck1 = "J♦ 7♥ 2♥ 3♣ 5♥ 5♦ 2♦ 3♦ Q♦ 9♠ A♣ 9♣ T♠ 7♦ J♥ 4♦ A♦ J♣ K♣ 9♥ T♦ 2♠ 6♣ 2♣ 6♦ 7♣ 8♣ K♠ 8♠ 6♠ 5♠ 6♥ Q♠ 5♣ Q♣ Q♥ 4♣ 3♠ A♠ 8♦ K♦ 9♦ 4♥ K♥ 8♥ T♥ 3♥ A♥ T♣ J♠ 7♠ 4♠"
func Deck1() *hand.Deck {
	return parseDeck(deck1Str)
}

func Deck2() *hand.Deck {
	return parseDeck(deck2Str)
}

func Deck3() *hand.Deck {
	return parseDeck(deck3Str)
}

func Deck4() *hand.Deck {
	return parseDeck(deck4Str)
}

func Deck5() *hand.Deck {
	return parseDeck(deck5Str)
}

func parseDeck(s string) *hand.Deck {
	cards := []hand.Card{}
	for _, cardStr := range strings.Split(s, " ") {
		temp := hand.AceSpades
		c := &temp
		if err := c.UnmarshalText([]byte(cardStr)); err != nil {
			panic(err)
		}
		cards = append(cards, *c)
	}
	return &hand.Deck{Cards: cards}
}

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
