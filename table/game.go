package table

import (
	"errors"

	"github.com/loganjspears/joker/hand"
	"github.com/loganjspears/joker/pot"
	"github.com/loganjspears/joker/util"
)

// A Game represents one of the different poker variations.
type Game int

const (
	// Holdem (also known as Texas hold'em) is a poker variation in which
	// players can combine two hole cards and five board cards to form the best
	// five card hand.  Holdem is typically played No Limit.
	Holdem Game = iota + 1

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

// Games returns all Games.
func Games() []Game {
	return []Game{Holdem, OmahaHi, OmahaHiLo, Razz, StudHi, StudHiLo}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (g Game) MarshalText() (text []byte, err error) {
	return []byte(g.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (g Game) UnmarshalText(text []byte) error {
	s := string(text)
	for _, gm := range Games() {
		if gm.String() == s {
			g = gm
			return nil
		}
	}
	return errors.New("table: game's unmarshaltext didn't find constant")
}

func (g Game) get() game {
	switch g {
	case Holdem:
		return holdem
	case OmahaHi:
		return omahaHi
	case OmahaHiLo:
		return omahaHiLo
	case StudHi:
		return studHi
	case Razz:
		return razz
	case StudHiLo:
		return studHiLo
	}
	panic("unreachable")
}

var (
	holdem game = &holdemGame{
		Split:   false,
		IsOmaha: false,
	}

	omahaHi game = &holdemGame{
		Split:   false,
		IsOmaha: true,
	}

	omahaHiLo game = &holdemGame{
		Split:   true,
		IsOmaha: true,
	}

	studHi game = &studGame{
		Split:  false,
		IsRazz: false,
	}

	razz game = &studGame{
		Split:  false,
		IsRazz: true,
	}

	studHiLo game = &studGame{
		Split:  true,
		IsRazz: false,
	}
)

type holeCards map[int][]*HoleCard

type game interface {
	NumOfRounds() int
	MaxSeats() int
	HoleCards(deck *hand.Deck, r round) []*HoleCard
	BoardCards(deck *hand.Deck, r round) []*hand.Card
	SplitPot() bool
	Sorting() hand.Sorting
	FormHighHand(holeCards []*hand.Card, boardCards []*hand.Card) *hand.Hand
	FormLowHand(holeCards []*hand.Card, boardCards []*hand.Card) *hand.Hand
	ForcedBet(holeCards holeCards, opts Config, r round, seat, relativePos int) int
	RoundStartSeat(holeCards holeCards, r round) int
	FixedLimit(opts Config, r round) int
}

type holdemGame struct {
	Split   bool
	IsOmaha bool
}

func (g *holdemGame) NumOfRounds() int {
	return 4
}

func (g *holdemGame) MaxSeats() int {
	return 10
}

func (g *holdemGame) HoleCards(deck *hand.Deck, r round) []*HoleCard {
	numOfCards := 2
	if g.IsOmaha {
		numOfCards = 4
	}
	switch r {
	case preflop:
		return holeCardsPopMulti(deck, Concealed, numOfCards)
	}
	return []*HoleCard{}
}

func (g *holdemGame) BoardCards(deck *hand.Deck, r round) []*hand.Card {
	switch r {
	case flop:
		return deck.PopMulti(3)
	case turn:
		return deck.PopMulti(1)
	case river:
		return deck.PopMulti(1)
	}
	return []*hand.Card{}
}

func (g *holdemGame) SplitPot() bool {
	return g.Split
}

func (g *holdemGame) Sorting() hand.Sorting {
	return hand.SortingHigh
}

func (g *holdemGame) FormHighHand(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
	if !g.IsOmaha {
		cards := append(board, holeCards...)
		return hand.New(cards)
	}

	opts := func(c *hand.Config) {}
	hands := omahaHands(holeCards, board, opts)
	hands = hand.Sort(hand.SortingHigh, hand.DESC, hands...)
	return hands[0]
}

func (g *holdemGame) FormLowHand(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
	if !g.IsOmaha {
		return nil
	}

	hands := omahaHands(holeCards, board, hand.AceToFiveLow)
	hands = hand.Sort(hand.SortingLow, hand.DESC, hands...)
	if hands[0].CompareTo(eightOrBetter) <= 0 {
		return hands[0]
	}
	return nil
}

func (g *holdemGame) ForcedBet(holeCards holeCards, opts Config, r round, seat, relativePos int) int {
	chips := 0
	if r != preflop {
		return chips
	}

	chips += opts.Stakes.Ante

	// reduce blind sizes if fixed limit
	smallBet := opts.Stakes.SmallBet
	bigBet := opts.Stakes.BigBet
	if opts.Limit == FixedLimit {
		smallBet /= 2
		bigBet /= 2
	}

	numOfPlayers := len(holeCards)
	if numOfPlayers == 2 {
		switch relativePos {
		case 0:
			chips += smallBet
		case 1:
			chips += bigBet
		}
	} else {
		switch relativePos {
		case 1:
			chips += smallBet
		case 2:
			chips += bigBet
		}
	}
	return chips
}

func (g *holdemGame) RoundStartSeat(holeCards holeCards, r round) int {
	numOfPlayers := len(holeCards)
	if r != preflop {
		return 1
	}
	switch numOfPlayers {
	case 2, 3:
		return 0
	}
	return 3
}

func (g *holdemGame) FixedLimit(opts Config, r round) int {
	switch r {
	case turn, river:
		return opts.Stakes.BigBet
	}
	return opts.Stakes.SmallBet
}

type studGame struct {
	Split  bool
	IsRazz bool
}

func (g *studGame) NumOfRounds() int {
	return 5
}

func (g *studGame) MaxSeats() int {
	return 8
}

func (g *studGame) HoleCards(deck *hand.Deck, r round) []*HoleCard {
	switch r {
	case thirdSt:
		cards := holeCardsPopMulti(deck, Concealed, 2)
		cards = append(cards, newHoleCard(deck.Pop(), Exposed))
		return cards
	case fourthSt:
		return holeCardsPopMulti(deck, Exposed, 1)
	case fifthSt:
		return holeCardsPopMulti(deck, Exposed, 1)
	case sixthSt:
		return holeCardsPopMulti(deck, Exposed, 1)
	case seventhSt:
		return holeCardsPopMulti(deck, Concealed, 1)
	}
	return []*HoleCard{}
}

// TODO: take into account running out of cards
func (g *studGame) BoardCards(deck *hand.Deck, r round) []*hand.Card {
	return []*hand.Card{}
}

func (g *studGame) SplitPot() bool {
	return g.Split
}

func (g *studGame) Sorting() hand.Sorting {
	if g.IsRazz {
		return hand.SortingLow
	}
	return hand.SortingHigh
}

func (g *studGame) FormHighHand(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
	cards := append(board, holeCards...)
	if g.IsRazz {
		return hand.New(cards, hand.AceToFiveLow)
	}
	return hand.New(cards)
}

func (g *studGame) FormLowHand(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
	cards := append(board, holeCards...)
	hand := hand.New(cards, hand.AceToFiveLow)
	if hand.CompareTo(eightOrBetter) <= 0 {
		return hand
	}
	return nil
}

func (g *studGame) ForcedBet(holeCards holeCards, opts Config, r round, seat, relativePos int) int {
	chips := 0
	if r != thirdSt {
		return chips
	}

	chips += opts.Stakes.Ante
	startSeat := g.RoundStartSeat(holeCards, r)
	if startSeat == seat {
		chips += opts.Stakes.SmallBet
	}

	return chips
}

func (g *studGame) RoundStartSeat(holeCards holeCards, r round) int {
	exposed := exposedCards(holeCards)
	f := func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		return hand.New(holeCards)
	}

	sorting := hand.SortingLow
	if (r != thirdSt && !g.IsRazz) || (r == thirdSt && g.IsRazz) {
		sorting = hand.SortingHigh
	}
	hands := pot.NewHands(exposed, []*hand.Card{}, f)
	hands = hands.WinningHands(sorting)

	for seat := range hands {
		return seat
	}
	panic("unreachable")
}

func (g *studGame) FixedLimit(opts Config, r round) int {
	switch r {
	case thirdSt, fourthSt:
		return opts.Stakes.SmallBet
	}
	return opts.Stakes.BigBet
}

func omahaHands(holeCards []*hand.Card, board []*hand.Card, opts func(*hand.Config)) []*hand.Hand {
	hands := []*hand.Hand{}
	selected := make([]*hand.Card, 2)
	for _, indexes := range util.Combinations(4, 2) {
		for j, i := range indexes {
			selected[j] = holeCards[i]
		}
		cards := append(board, selected...)
		hands = append(hands, hand.New(cards, opts))
	}
	return hands
}

func exposedCards(holeCards map[int][]*HoleCard) map[int][]*hand.Card {
	exposed := map[int][]*hand.Card{}
	for seat, hCards := range holeCards {
		eCards := []*hand.Card{}
		for _, hc := range hCards {
			if hc.Visibility == Exposed {
				eCards = append(eCards, hc.Card)
			}
		}
		exposed[seat] = eCards
	}
	return exposed
}

type round int

const (
	preflop round = 0
	flop    round = 1
	turn    round = 2
	river   round = 3

	thirdSt   round = 0
	fourthSt  round = 1
	fifthSt   round = 2
	sixthSt   round = 3
	seventhSt round = 4
)

var (
	eightOrBetter = hand.New([]*hand.Card{
		hand.EightSpades,
		hand.SevenSpades,
		hand.SixSpades,
		hand.FiveSpades,
		hand.FourSpades,
	}, hand.AceToFiveLow)
)
