package table

import (
	"github.com/SyntropyDev/joker/hand"
	"github.com/SyntropyDev/joker/util"
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

// String returns the Game's name.
func (g Game) String() string {
	return g.getGameType().name
}

func (g Game) getGameType() game {
	switch g {
	case Holdem:
		return holdem
	case OmahaHi:
		return omahaHi
	case OmahaHiLo:
		return omahaHiLo
	case Razz:
		return razz
	case StudHi:
		return studHi
	case StudHiLo:
		return studHiLo
	}
	panic("unreachable")
}

type game struct {
	name           string
	numOfRounds    int
	maxSeats       int
	winType        winType
	holeCards      func(deck hand.Deck, r round) []*HoleCard
	boardCards     func(deck hand.Deck, r round) []*hand.Card
	highHand       handCreationFunc
	lowHand        handCreationFunc
	forcedBet      func(holeCards map[int][]*HoleCard, limit Limit, stakes Stakes, r round, seat, relativePos int) int
	roundStartSeat func(holeCards map[int][]*HoleCard, r round, numOfPlayers int) int
	fixedLimit     func(stakes Stakes, r round) int
}

var (
	holdem = game{
		name:        "Holdem",
		numOfRounds: holdemNumOfRounds,
		maxSeats:    holdemMaxSeats,
		winType:     winHigh,
		holeCards: func(deck hand.Deck, r round) []*HoleCard {
			switch r {
			case preflop:
				return holeCardsPopMulti(deck, Concealed, 2)
			}
			return []*HoleCard{}
		},
		boardCards: holdemBoardFunc,
		highHand: func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
			cards := append(board, holeCards...)
			return hand.New(cards)
		},
		forcedBet:      holdemBlinds,
		roundStartSeat: holdemStartSeatFunc,
		fixedLimit: func(stakes Stakes, r round) int {
			switch r {
			case preflop, flop:
				return stakes.SmallBet
			case turn, river:
				return stakes.BigBet
			}
			panic("should not get here")
		},
	}

	omahaHi = game{
		name:        "Omaha Hi",
		numOfRounds: holdemNumOfRounds,
		maxSeats:    holdemMaxSeats,
		winType:     winHigh,
		holeCards: func(deck hand.Deck, r round) []*HoleCard {
			switch r {
			case preflop:
				return holeCardsPopMulti(deck, Concealed, 4)
			}
			return []*HoleCard{}
		},
		boardCards:     holdemBoardFunc,
		highHand:       omahaHighHand,
		forcedBet:      holdemBlinds,
		roundStartSeat: holdemStartSeatFunc,
	}

	omahaHiLo = game{
		name:        "Omaha Hi/Lo",
		numOfRounds: holdemNumOfRounds,
		maxSeats:    holdemMaxSeats,
		winType:     winHighLow,
		holeCards: func(deck hand.Deck, r round) []*HoleCard {
			switch r {
			case preflop:
				return holeCardsPopMulti(deck, Concealed, 4)
			}
			return []*HoleCard{}
		},
		boardCards: holdemBoardFunc,
		highHand:   omahaHighHand,
		lowHand: func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
			hands := omahaHands(holeCards, board, hand.AceToFiveLow)
			hands = hand.Sort(hand.SortingLow, hand.DESC, hands...)
			if hands[0].CompareTo(eightOrBetter) <= 0 {
				return hands[0]
			}
			return nil
		},
		forcedBet:      holdemBlinds,
		roundStartSeat: holdemStartSeatFunc,
	}

	razz = game{
		name:        "Razz",
		numOfRounds: studNumOfRounds,
		maxSeats:    studMaxSeats,
		winType:     winLow,
		holeCards:   studHoleCardFunc,
		boardCards:  studBoardFunc,
		lowHand: func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
			return hand.New(holeCards, hand.AceToFiveLow)
		},
		forcedBet:      studBringIn(winHigh),
		roundStartSeat: studRoundStartSeat(winHigh, winLow),
	}

	studHi = game{
		name:        "Stud Hi",
		numOfRounds: studNumOfRounds,
		maxSeats:    studMaxSeats,
		winType:     winHigh,
		holeCards:   studHoleCardFunc,
		boardCards:  studBoardFunc,
		highHand: func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
			return hand.New(holeCards)
		},
		forcedBet:      studBringIn(winLow),
		roundStartSeat: studRoundStartSeat(winLow, winHigh),
	}

	studHiLo = game{
		name:        "Stud Hi/Lo",
		numOfRounds: studNumOfRounds,
		maxSeats:    studMaxSeats,
		winType:     winHighLow,
		holeCards:   studHoleCardFunc,
		boardCards:  studBoardFunc,
		highHand: func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
			return hand.New(holeCards)
		},
		lowHand: func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
			hand := hand.New(holeCards, hand.AceToFiveLow)
			if hand.CompareTo(eightOrBetter) <= 0 {
				return hand
			}
			return nil
		},
		forcedBet:      studBringIn(winLow),
		roundStartSeat: studRoundStartSeat(winLow, winHigh),
	}
)

func holdemBlinds(
	holeCards map[int][]*HoleCard,
	limit Limit,
	stakes Stakes,
	r round,
	seat, relativePos int) int {

	chips := 0
	if r != preflop {
		return chips
	}

	chips += stakes.Ante

	// reduce blind sizes if fixed limit
	smallBet := stakes.SmallBet
	bigBet := stakes.BigBet
	if limit == FixedLimit {
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

func omahaHands(holeCards []*hand.Card, board []*hand.Card, opts func(*hand.Config)) []*hand.Hand {
	hands := []*hand.Hand{}
	for _, indexes := range util.Combinations(4, 2) {
		selected := []*hand.Card{}
		for _, i := range indexes {
			selected = append(selected, holeCards[i])
		}
		cards := append(board, selected...)
		hands = append(hands, hand.New(cards, opts))
	}
	return hands
}

// TODO fix bring in sizes for fixed limit
func studBringIn(winType winType) func(map[int][]*HoleCard, Limit, Stakes, round, int, int) int {
	return func(holeCards map[int][]*HoleCard, limit Limit, stakes Stakes, r round, seat, relativePos int) int {
		chips := 0
		if r != thirdSt {
			return chips
		}

		chips += stakes.Ante
		f := studRoundStartSeat(winType, winHigh)
		startSeat := f(holeCards, r, len(holeCards))
		if startSeat == seat {
			chips += stakes.SmallBet
		}

		return chips
	}
}

func studRoundStartSeat(w1 winType, w2 winType) func(map[int][]*HoleCard, round, int) int {
	return func(holeCards map[int][]*HoleCard, r round, numOfPlayers int) int {
		exposed := exposedCards(holeCards)
		f := func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
			return hand.New(holeCards)
		}

		hands := newHands(exposed, []*hand.Card{}, f)
		if r == thirdSt {
			hands = hands.WinningHands(w1)
		} else {
			hands = hands.WinningHands(w2)
		}

		for seat := range hands {
			return seat
		}
		panic("should not get here")
	}
}

func exposedCards(holeCards map[int][]*HoleCard) map[int][]*HoleCard {
	exposed := map[int][]*HoleCard{}
	for seat, hCards := range holeCards {
		eCards := []*HoleCard{}
		for _, hc := range hCards {
			if hc.Visibility == Exposed {
				eCards = append(eCards, hc)
			}
		}
		exposed[seat] = eCards
	}
	return exposed
}

type winType int

const (
	winHigh winType = iota + 1
	winLow
	winHighLow
)

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

func holdemRounds() []round {
	return []round{preflop, flop, turn, river}
}

func studRounds() []round {
	return []round{thirdSt, fourthSt, fifthSt, sixthSt, seventhSt}
}

const (
	holdemNumOfRounds = 4
	holdemMaxSeats    = 10

	studNumOfRounds = 5
	studMaxSeats    = 8
)

var (
	holdemBoardFunc = func(deck hand.Deck, r round) []*hand.Card {
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

	omahaHighHand = func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand {
		opts := func(c *hand.Config) {}
		hands := omahaHands(holeCards, board, opts)
		hands = hand.Sort(hand.SortingHigh, hand.DESC, hands...)
		return hands[0]
	}

	studHoleCardFunc = func(deck hand.Deck, r round) []*HoleCard {
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

	studBoardFunc = func(deck hand.Deck, r round) []*hand.Card {
		return []*hand.Card{}
	}

	holdemStartSeatFunc = func(holeCards map[int][]*HoleCard, r round, numOfPlayers int) int {
		if r != preflop {
			return 1
		}

		switch numOfPlayers {
		case 2, 3:
			return 0
		default:
			return 3
		}
	}

	eightOrBetter = hand.New([]*hand.Card{
		hand.EightSpades,
		hand.SevenSpades,
		hand.SixSpades,
		hand.FiveSpades,
		hand.FourSpades,
	}, hand.AceToFiveLow)
)
