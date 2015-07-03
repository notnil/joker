package game

import (
	"github.com/loganjspears/joker/hand"
	"github.com/loganjspears/joker/util"
)

// Game provides functionality required of all poker games.
type Game interface {
	Rounds() int
	MaxSeats() int
	Showdown(holeCards map[int][]*hand.Card, board []*hand.Card) Results
}

// Type represents a poker variant.
type Type int

const (
	// Holdem (also known as Texas hold'em) is a poker variation in which
	// players can combine two hole cards and five board cards to form the best
	// five card hand.  Holdem is typically played No Limit.
	Holdem Type = iota + 1

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

// Rounds return the number of rounds for the type.
func (t Type) Rounds() int {
	switch t {
	case Holdem, OmahaHi, OmahaHiLo:
		return 4
	case Razz, StudHi, StudHiLo:
		return 5
	}
	return 0
}

// MaxSeats return the maximum number of seats for the type.
func (t Type) MaxSeats() int {
	switch t {
	case Holdem, OmahaHi, OmahaHiLo:
		return 10
	case Razz, StudHi, StudHiLo:
		return 8
	}
	return 0
}

// Showdown return the showdown results from the hole cards and board. holecards
// should be a mapping of seat to hole cards.
func (t Type) Showdown(holeCards map[int][]*hand.Card, board []*hand.Card) Results {
	switch t {
	case Holdem, StudHi:
		return newHands(holeCards, board, holdemFunc).
			winningHands(hand.SortingHigh).
			results(hand.SortingHigh)
	case OmahaHi:
		return newHands(holeCards, board, omahaHiFunc).
			winningHands(hand.SortingHigh).
			results(hand.SortingHigh)
	case OmahaHiLo:
		return newHands(holeCards, board, omahaHiFunc).
			winningHands(hand.SortingHigh).
			results(hand.SortingHigh).merge(
			newHands(holeCards, board, omahaLowFunc).
				winningHands(hand.SortingLow).
				results(hand.SortingLow))
	case Razz:
		return newHands(holeCards, board, razzFunc).
			winningHands(hand.SortingLow).
			results(hand.SortingHigh)
	case StudHiLo:
		return newHands(holeCards, board, holdemFunc).
			winningHands(hand.SortingHigh).
			results(hand.SortingHigh).merge(
			newHands(holeCards, board, studLoFunc).
				winningHands(hand.SortingLow).
				results(hand.SortingLow))
	}
	return nil
}

// Types returns all Types.
func Types() []Type {
	return []Type{Holdem, OmahaHi, OmahaHiLo, Razz, StudHi, StudHiLo}
}

var (
	holdemFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		cards := append(board, holeCards...)
		return hand.New(cards)
	}
	omahaHiFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		opts := func(c *hand.Config) {}
		hands := omahaHands(holeCards, board, opts)
		hands = hand.Sort(hand.SortingHigh, hand.DESC, hands...)
		return hands[0]
	}
	omahaLowFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		hands := omahaHands(holeCards, board, hand.AceToFiveLow)
		hands = hand.Sort(hand.SortingLow, hand.DESC, hands...)
		if hands[0].CompareTo(eightOrBetter) <= 0 {
			return hands[0]
		}
		return nil
	}
	razzFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		cards := append(board, holeCards...)
		return hand.New(cards, hand.AceToFiveLow)
	}
	studLoFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		cards := append(board, holeCards...)
		hand := hand.New(cards, hand.AceToFiveLow)
		if hand.CompareTo(eightOrBetter) <= 0 {
			return hand
		}
		return nil
	}
	eightOrBetter = hand.New([]*hand.Card{
		hand.EightSpades,
		hand.SevenSpades,
		hand.SixSpades,
		hand.FiveSpades,
		hand.FourSpades,
	}, hand.AceToFiveLow)
)

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
