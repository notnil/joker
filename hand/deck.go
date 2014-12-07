package hand

import (
	"math/rand"
	"strings"
	"time"
)

// Deck is a slice of cards used for dealing
type Deck struct {
	Cards []*Card
}

// Pop removes a card from the deck and returns it.  Pop
// panics if no cards are available.
func (d *Deck) Pop() *Card {
	last := len(d.Cards) - 1
	card := d.Cards[last]
	d.Cards = d.Cards[:last]
	return card
}

// PopMulti calls the Pop function on n number of cards.
func (d *Deck) PopMulti(n int) []*Card {
	cards := []*Card{}
	for i := 0; i < n; i++ {
		cards = append(cards, d.Pop())
	}
	return cards
}

// String implements the fmt.Stringer interface
func (d *Deck) String() string {
	s := []string{}
	for _, c := range d.Cards {
		s = append(s, c.String())
	}
	return strings.Join(s, ",")
}

// MarshalText implements the encoding.TextMarshaler interface
func (d *Deck) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (d *Deck) UnmarshalText(text []byte) error {
	strs := strings.Split(string(text), ",")
	cards := make([]*Card, len(strs))
	for i, s := range strs {
		card := &Card{}
		if err := card.UnmarshalText([]byte(s)); err != nil {
			return err
		}
		cards[i] = card
	}
	d.Cards = cards
	return nil
}

// Dealer provides a way to generate new decks.
type Dealer interface {
	Deck() *Deck
}

// NewDealer returns a dealer that generates shuffled decks.
func NewDealer() Dealer {
	return dealer{}
}

type dealer struct{}

func (d dealer) Deck() *Deck {
	cards := shuffleCards(Cards())
	return &Deck{Cards: cards}
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
