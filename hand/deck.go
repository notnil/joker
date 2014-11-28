package hand

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
// FromCards forms a deck from its remaining cards and discards.  It
// is required for serialization.
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
	Cards() []*Card
	Discard(cards ...*Card)
	Discards() []*Card
	FromCards(cards, discards []*Card) Deck
	Pop() *Card
	PopMulti(n int) []*Card
	Reset()
}

// NewDeck returns a new deck of with 52 shuffled cards.
func NewDeck() *ShuffledDeck {
	cards := shuffleCards(Cards())
	return &ShuffledDeck{cards: cards, discards: []*Card{}}
}

// EmptyDeck returns an deck with no cards
func EmptyDeck() *ShuffledDeck {
	return &ShuffledDeck{cards: []*Card{}, discards: []*Card{}}
}

// ShuffledDeck implements the Deck interface
type ShuffledDeck struct {
	cards    []*Card
	discards []*Card
}

func (d *ShuffledDeck) Cards() []*Card {
	cards := make([]*Card, len(d.cards))
	copy(cards, d.cards)
	return cards
}

// Discard adds the given cards to a reservoir of cards that can be
// used if the deck is empty.
func (d *ShuffledDeck) Discard(cards ...*Card) {
	d.discards = append(d.discards, cards...)
}

func (d *ShuffledDeck) Discards() []*Card {
	cards := make([]*Card, len(d.discards))
	copy(cards, d.discards)
	return cards
}

// FromCards forms a deck from its remaining cards and discards.  It
// is required for serialization.
func (d *ShuffledDeck) FromCards(cards, discards []*Card) Deck {
	return &ShuffledDeck{
		cards:    cards,
		discards: discards,
	}
}

// Pop removes a card from the deck and returns it.  If no card is
// available then the discards should be shuffled and reused.
func (d *ShuffledDeck) Pop() *Card {
	// TODO need to utilize discards if cards required is greater than 52.
	last := len(d.cards) - 1
	cards, card := d.cards[:last], d.cards[last]
	d.cards = cards
	return card
}

// PopMulti calls the Pop function on n number of cards.
func (d *ShuffledDeck) PopMulti(n int) []*Card {
	cards := []*Card{}
	for i := 0; i < n; i++ {
		cards = append(cards, d.Pop())
	}
	return cards
}

// Reset restores the deck for a new hand with 52 dealable cards and
// no discards.
func (d *ShuffledDeck) Reset() {
	d.cards = NewDeck().PopMulti(52)
	d.discards = []*Card{}
}

type deckJSON struct {
	Cards    []*Card
	Discards []*Card
}

// MarshalJSON implements the json.Marshaler interface
func (d *ShuffledDeck) MarshalJSON() ([]byte, error) {
	m := &deckJSON{
		Cards:    d.cards,
		Discards: d.discards,
	}
	return json.Marshal(&m)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *ShuffledDeck) UnmarshalJSON(data []byte) error {
	m := &deckJSON{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	d.cards = m.Cards
	d.discards = m.Discards
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
