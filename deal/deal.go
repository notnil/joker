package deal

import (
	"errors"

	"github.com/notnil/joker/hand"
	"github.com/notnil/joker/pot"
)

type Deal struct {
	config    Config
	pot       *pot.Pot
	round     int
	board     []hand.Card
	holeCards map[int][]hand.Card
	payouts   []*pot.Payout
	hands     map[int]*hand.Hand
}

type Config struct {
	Variant Variant
	Deck    *hand.Deck
	Button  int
	Stacks  map[int]int
	Blinds  []int
	Ante    int
}

func New(c Config) *Deal {
	d := &Deal{
		config:    c,
		board:     []hand.Card{},
		holeCards: map[int][]hand.Card{},
		pot:       nil,
		round:     0,
	}
	d.config.Variant.getUpdater().Update(d)
	return d
}

func (d *Deal) Variant() Variant {
	return d.config.Variant
}

func (d *Deal) Deck() *hand.Deck {
	return &hand.Deck{Cards: append([]hand.Card{}, d.config.Deck.Cards...)}
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
	case pot.AllIn:
		if err := d.pot.AllIn(); err != nil {
			return err
		}
	default:
		return errors.New("deal: unknown action")
	}
	if d.pot.SeatToAct() == nil {
		// payout player and end deal if there is only one player
		if payout := d.pot.Uncontested(); payout != nil {
			d.payouts = []*pot.Payout{payout}
			return nil
		}
		d.round++
		d.config.Variant.getUpdater().Update(d)
		// if everyone is all-in continue till payouts
		if d.pot.SeatToAct() == nil {
			for d.payouts == nil {
				d.round++
				d.config.Variant.getUpdater().Update(d)
			}
		}
	}
	return nil
}
