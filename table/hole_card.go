package table

import "github.com/SyntropyDev/joker/hand"

// CardVisibility indicates a HoleCard's visibility to other players
type CardVisibility string

const (
	// Concealed indicates a HoleCard should be hidden from other players
	Concealed CardVisibility = "Concealed"

	// Exposed indicates a HoleCard should be shown to other players
	Exposed CardVisibility = "Exposed"
)

// A HoleCard represents a card that has been dealt to a player.
type HoleCard struct {
	// Card is the HoleCard's card value.
	Card *hand.Card `json:"card"`

	// Visibility is the HoleCard's CardVisibility.
	Visibility CardVisibility `json:"visibility"`
}

// String returns the Card's String() result
func (c *HoleCard) String() string {
	return c.Card.String()
}

func concealedCard() *HoleCard {
	return &HoleCard{nil, Concealed}
}

func newHoleCard(card *hand.Card, visibility CardVisibility) *HoleCard {
	return &HoleCard{card, visibility}
}

func holeCardsPopMulti(d hand.Deck, v CardVisibility, n int) []*HoleCard {
	cards := d.PopMulti(n)
	holeCards := []*HoleCard{}
	for _, c := range cards {
		holeCards = append(holeCards, newHoleCard(c, v))
	}
	return holeCards
}

func cardsFromHoleCards(holeCards []*HoleCard) []*hand.Card {
	cards := []*hand.Card{}
	for _, hc := range holeCards {
		cards = append(cards, hc.Card)
	}
	return cards
}

func tableViewOfHoleCards(holeCards []*HoleCard) []*HoleCard {
	hCards := []*HoleCard{}
	for _, hc := range holeCards {
		if hc.Visibility == Exposed {
			hCards = append(hCards, hc)
		} else if hc.Visibility == Concealed {
			hCards = append(hCards, &HoleCard{Card: nil, Visibility: Concealed})
		}
	}
	return hCards
}
