package table

import "github.com/SyntropyDev/joker/hand"

type tableHands map[int]*hand.Hand
type handCreationFunc func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand

func newHands(
	seatToHoleCards map[int][]*HoleCard,
	board []*hand.Card,
	f handCreationFunc) tableHands {

	hands := map[int]*hand.Hand{}
	for seat, holeCards := range seatToHoleCards {
		hCards := cardsFromHoleCards(holeCards)
		if f != nil {
			hands[seat] = f(hCards, board)
		}
	}
	return tableHands(hands)
}

func (hands tableHands) HandsForSeats(seats []int) tableHands {
	newHands := map[int]*hand.Hand{}
	for seat, hand := range hands {
		found := false
		for _, s := range seats {
			found = found || s == seat
		}
		if found {
			newHands[seat] = hand
		}
	}
	return newHands
}

func (hands tableHands) WinningHands(winType winType) tableHands {
	// copy all eligible hands (for Stud8 & Omaha8)
	handsMapCopy := map[int]*hand.Hand{}
	for seat, hand := range hands {
		if hand != nil {
			handsMapCopy[seat] = hand
		}
	}
	handsCopy := tableHands(handsMapCopy)
	if len(handsCopy) == 0 {
		return handsCopy
	}

	s := handsCopy.slice()
	sorting := hand.SortingHigh
	if winType == winLow {
		sorting = hand.SortingLow
	}
	s = hand.Sort(sorting, hand.DESC, s...)
	best := s[0]

	selected := map[int]*hand.Hand{}
	for seat, hand := range handsCopy {
		if best.CompareTo(hand) == 0 {
			selected[seat] = hand
		}
	}
	return tableHands(selected)
}

func (hands tableHands) slice() []*hand.Hand {
	s := []*hand.Hand{}
	for _, h := range hands {
		s = append(s, h)
	}
	return s
}
