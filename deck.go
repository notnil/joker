package joker

import (
	"encoding/json"
	"math/rand"
	"time"
)

// Deck is the interface that manages cards in a perserved order.
//
// Discard adds the given cards to a reservoir of cards that can be
// used if the deck is empty.
//
// Len return the number of cards remaining in the deck.
//
// Pop removes a card from the deck and returns it.  If no card is
// available then the discards should be shuffled and reused.
//
// PopMulti calls the Pop function on n number of cards.
//
// Reset restores the deck for a new hand with 52 dealable cards and
// no discards.
type Deck interface {
	json.Marshaler
	json.Unmarshaler

	Discard(cards ...*Card)
	Len() int
	Pop() *Card
	PopMulti(n int) []*Card
	Reset()
}

// NewDeck returns a new deck of with 52 shuffled cards.
func NewDeck() Deck {
	cards := shuffleCards(Cards())
	return &defaultDeck{cards: cards, discards: []*Card{}}
}

type defaultDeck struct {
	cards    []*Card
	discards []*Card
}

func (d *defaultDeck) Discard(cards ...*Card) {
	d.discards = append(d.discards, cards...)
}

func (d *defaultDeck) Len() int {
	return len(d.cards)
}

// TODO need to utilize discards if cards required is greater than 52.
func (d *defaultDeck) Pop() *Card {
	last := len(d.cards) - 1
	cards, card := d.cards[:last], d.cards[last]
	d.cards = cards
	return card
}

func (d *defaultDeck) PopMulti(n int) []*Card {
	cards := []*Card{}
	for i := 0; i < n; i++ {
		cards = append(cards, d.Pop())
	}
	return cards
}

func (d *defaultDeck) Reset() {
	d.cards = NewDeck().PopMulti(52)
	d.discards = []*Card{}
}

func (d *defaultDeck) MarshalJSON() ([]byte, error) {
	m := map[string][]*Card{
		"cards":    d.cards,
		"discards": d.discards,
	}
	return json.Marshal(&m)
}

func (d *defaultDeck) UnmarshalJSON(data []byte) error {
	m := map[string][]*Card{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	d.cards = m["cards"]
	d.discards = m["discards"]
	return nil
}

func shuffleCards(cards []*Card) []*Card {
	rand.Seed(time.Now().UTC().UnixNano())
	dest := []*Card{}
	perm := rand.Perm(len(cards))
	for _, v := range perm {
		dest = append(dest, cards[v])
	}
	return dest
}
