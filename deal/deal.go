package deal

import (
	"errors"

	"github.com/notnil/joker/hand"
	"github.com/notnil/joker/pot"
)

type Deal struct {
	variant        Variant
	deck           *hand.Deck
	board          []hand.Card
	holeCards      map[int][]hand.Card
	pot            *pot.Pot
	round          int
	startingStacks map[int]int
	button         int
	payouts        []*pot.Payout
	hands          map[int]*hand.Hand
}

func New(v Variant, deck *hand.Deck, startingStacks map[int]int, button int) *Deal {
	d := &Deal{
		variant:        v,
		deck:           deck,
		board:          []hand.Card{},
		holeCards:      map[int][]hand.Card{},
		pot:            nil,
		round:          0,
		startingStacks: startingStacks,
		button:         button,
	}
	d.variant.getUpdater().Update(d)
	return d
}

func (d *Deal) Variant() Variant {
	return d.variant
}

func (d *Deal) Deck() *hand.Deck {
	return &hand.Deck{Cards: append([]hand.Card{}, d.deck.Cards...)}
}

func (d *Deal) Board() []hand.Card {
	return append([]hand.Card{}, d.board...)
}

func (d *Deal) HoleCards() map[int][]hand.Card {
	cp := map[int][]hand.Card{}
	for k, v := range d.holeCards {
		cp[k] = append([]hand.Card{}, v...)
	}
	return cp
}

func (d *Deal) Pot() *pot.Pot {
	return d.pot
}

func (d *Deal) Round() int {
	return d.round
}

func (d *Deal) Payouts() []*pot.Payout {
	if d.payouts == nil {
		return nil
	}
	return append([]*pot.Payout{}, d.payouts...)
}

func (d *Deal) Hands() map[int]*hand.Hand {
	return d.hands
}

func (d *Deal) Action(a pot.Action, chips int) error {
	if d.payouts != nil {
		return errors.New("deal: no actions can be taken after payout")
	}
	switch a {
	case pot.Fold:
		if err := d.pot.Fold(); err != nil {
			return err
		}
	case pot.Check:
		if err := d.pot.Check(); err != nil {
			return err
		}
	case pot.Call:
		if err := d.pot.Call(); err != nil {
			return err
		}
	case pot.Bet:
		if err := d.pot.Bet(chips); err != nil {
			return err
		}
	case pot.Raise:
		if err := d.pot.Raise(chips); err != nil {
			return err
		}
	default:
		return errors.New("deal: unknown action")
	}
	if d.pot.SeatToAct() == nil {
		// TODO check if pot is uncontested or everyone is all-in
		d.round++
		d.variant.getUpdater().Update(d)
	}
	return nil
}
