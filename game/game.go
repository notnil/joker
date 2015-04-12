package game

import (
	"github.com/loganjspears/joker/hand"
	"github.com/loganjspears/joker/util"
)

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

type Limit int

const (
	FixedLimit Limit = iota
	PotLimit
	NoLimit
)

type Stakes struct {
	SmallBet int
	BigBet   int
	Ante     int
}

// CardVisibility indicates a HoleCard's visibility to other players
type CardVisibility string

const (
	// Concealed indicates a HoleCard should be hidden from other players
	Concealed CardVisibility = "Concealed"

	// Exposed indicates a HoleCard should be shown to other players
	Exposed CardVisibility = "Exposed"
)

type Showdowner interface {
	Showdown(holeCards map[int][]*hand.Card, board []*hand.Card) Results
}

type Game interface {
	Showdowner
	NumOfRounds() int
	MaxSeats() int
	CardsForRound(r Round) (holeCards map[CardVisibility]int, numOfBoard int)
	RoundStartSeat(r Round, numOfPlayers int) int
	ForcedBets(r Round, stakes Stakes, numOfPlayers int) map[int]int
}

var (
	Holdem    Game = &holdemGame{holdemType: holdem}
	OmahaHi   Game = &holdemGame{holdemType: omahaHi}
	OmahaHiLo Game = &holdemGame{holdemType: holdem}
)

type holdemType int

const (
	holdem holdemType = iota
	omahaHi
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

func (g *holdemGame) CardsForRound(r Round) (holeCards map[CardVisibility]int, numOfBoard int) {
	switch r {
	case Preflop:
		holeCards[Concealed] = 2
		numOfBoard = 0
	case Flop:
		numOfBoard = 3
	case Turn:
		numOfBoard = 1
	case River:
		numOfBoard = 1
	}
	return
}

func (g *holdemGame) RoundStartSeat(r Round, numOfPlayers int) int {
	if r != Preflop {
		return 1
	}
	switch numOfPlayers {
	case 2, 3:
		return 0
	}
	return 3
}

func (g *holdemGame) ForcedBets(r Round, stakes Stakes, numOfPlayers int) (chips map[int]int) {
	if r != Preflop {
		return chips
	}

	for i := 0; i < numOfPlayers; i++ {
		chips[i] = stakes.Ante
	}

	if numOfPlayers == 2 {
		chips[0] = stakes.SmallBet
		chips[1] = stakes.BigBet
	} else {
		chips[1] = stakes.SmallBet
		chips[2] = stakes.BigBet
	}
	return
}

func (g *holdemGame) Showdown(holeCards map[int][]*hand.Card, board []*hand.Card) Results {
	if g.holdemType == omahaHiLo {
		highHands := newHands(holeCards, board, omahaHiHandFunc).winningHands(hand.SortingHigh)
		lowHands := newHands(holeCards, board, omahaLowHandFunc).winningHands(hand.SortingLow)
		results := highHands.formResults(WonHigh, SplitHigh)
		if len(lowHands) > 0 {
			lowResults := lowHands.formResults(WonLow, SplitLow)
			results.merge(lowResults)
		}
		return results
	}
	f := holdemHandFunc
	if g.holdemType == omahaHi {
		f = omahaHiHandFunc
	}
	return newHands(holeCards, board, f).
		winningHands(hand.SortingHigh).
		formResults(WonHigh, SplitHigh)
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

var (
	holdemHandFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		cards := append(holeCards, board...)
		return hand.New(cards)
	}
	omahaHiHandFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		opts := func(c *hand.Config) {}
		hands := omahaHands(holeCards, board, opts)
		hands = hand.Sort(hand.SortingHigh, hand.DESC, hands...)
		return hands[0]
	}
	omahaLowHandFunc = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		hands := omahaHands(holeCards, board, hand.AceToFiveLow)
		hands = hand.Sort(hand.SortingLow, hand.DESC, hands...)
		if hands[0].CompareTo(eightOrBetter) <= 0 {
			return hands[0]
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
