package deal

import (
	"errors"

	"github.com/notnil/joker/hand"
	"github.com/notnil/joker/pot"
)

// A Variant represents one of the different poker variations.
type Variant int

const (
	// Holdem (also known as Texas hold'em) is a poker variation in which
	// players can combine two hole cards and five board cards to form the best
	// five card hand.  Holdem is typically played No Limit.
	Holdem Variant = iota

	// OmahaHi (also known as simply Omaha) is a poker variation with four
	// hole cards and five board cards.  The best combination of two hole cards
	// and three board cards is used to determine the best hand.  OmahaHi is
	// typically played Pot Limit.
	OmahaHi

	// OmahaHiLo (also known as Omaha/8) is a version of Omaha where the
	// high hand can split the pot with the low hand if one qualifies.  The low
	// hand must be "eight or better" meaning that it must have or be below an
	// eight high.  OmahaHiLo is usually played Pot Limit.
	OmahaHiLo

	// Razz is a stud game in which players combine three concealed and four
	// exposed hole cards to form the lowest hand.  In Razz, aces are low and
	// straights and flushes don't count.  Razz is typically played Fixed Limit.
	Razz

	// StudHi (also known as 7 Card Stud) is a stud game in which players combine
	// three concealed and four exposed hole cards to form the best hand. StudHi
	// is typically played Fixed or Pot Limit.
	StudHi

	// StudHiLo (also known as Stud8) is a version of Stud where the high hand can
	// split the pot with the low hand if one qualifies. The low hand must be
	// "eight or better" meaning that it must have or be below an eight high.
	// StudHiLo is typically played Fixed or Pot Limit.
	StudHiLo
)

// Variants returns all Games.
func Variants() []Variant {
	return []Variant{Holdem, OmahaHi, OmahaHiLo, Razz, StudHi, StudHiLo}
}

var (
	variantStrs = []string{"Holdem", "OmahaHi", "OmahaHiLo", "Razz", "StudHi", "StudHiLo"}
)

func (v Variant) String() string {
	return variantStrs[v]
}

// MarshalText implements the encoding.TextMarshaler interface.
func (v Variant) MarshalText() (text []byte, err error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (v *Variant) UnmarshalText(text []byte) error {
	s := string(text)
	for _, vt := range Variants() {
		if vt.String() == s {
			*v = vt
			return nil
		}
	}
	return errors.New("game: game unmarshaltext didn't find constant")
}

func (v Variant) getUpdater() updater {
	return holdemUpdater{}
}

type updater interface {
	Update(d *Deal)
}

type holdemUpdater struct{}

const (
	preflop = iota
	flop
	turn
	river
	showdown
)

// TODO deal with hands that finish before the end
func (holdemUpdater) Update(d *Deal) {
	switch d.round {
	case preflop:
		d.pot = pot.New(d.startingStacks, d.button, pot.Blinds([]int{1, 2}))
		for _, seat := range d.pot.Seats() {
			d.holeCards[seat.Pos] = d.deck.PopMulti(2)
		}
	case flop:
		d.pot.NextRound()
		d.board = d.deck.PopMulti(3)
	case turn, river:
		d.pot.NextRound()
		d.board = append(d.board, d.deck.Pop())
	case showdown:
		r := &ranker{e: holdemHandEvaluator{}, s: hand.SortingHigh, o: hand.DESC}
		d.hands = r.hands(d.holeCards, d.board)
		d.payouts = d.pot.Payout(r.rank(d.hands), nil)
	}
}
