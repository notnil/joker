package game

import "github.com/loganjspears/joker/hand"

type Round int

const (
	Preflop Round = 0
	Flop    Round = 1
	Turn    Round = 2
	River   Round = 3

	ThirdSt   Round = 0
	FourthSt  Round = 1
	FifthSt   Round = 2
	SixthSt   Round = 3
	SeventhSt Round = 4
)

type Game interface {
	NumOfRounds() int
	MaxSeats() int
	CardsForRound(r Round) (numOfHole, numOfBoard int)
	Payout(holeCards map[int][]*hand.Card, boardCards []*hand.Card, pot *Pot, button int) Results
	ForcedBets(holeCards map[int][]*Holecard, opts Config, r Round) map[int]int
	RoundStartSeat(holeCards map[int][]*Holecard, r Round) int
}

type holdemType int

const (
	holdem holdemType = iota
	omaha
	omahaHiLo
)

type holdemGame struct {
	holdemType holdemType
}

func (g *holdemGame) NumOfRounds() int {
	return int(River) + 1
}
func (g *holdemGame) MaxSeats() int {
	return 10
}
func (g *holdemGame) CardsForRound(r Round) (numOfHole, numOfBoard int) {
	switch r {
	case Preflop:
		return 2, 0
	case Flop:
		return 0, 3
	case Turn:
		return 0, 1
	case River:
		return 0, 1
	}
}

func (g *holdemGame) Payout(holeCards map[int][]*hand.Card, boardCards []*hand.Card, pot *Pot, button int) Results {
	if holdemType == holdem {

	}
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
