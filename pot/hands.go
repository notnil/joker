package pot

import "github.com/loganjspears/joker/hand"

// HandCreationFunc represents the function signature for creating a hand.
type HandCreationFunc func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand

// Hands represents a map of seats to hands
type Hands map[int]*hand.Hand

// NewHands forms hands from seat's hole cards and the board.
func NewHands(holeCards map[int][]*hand.Card, board []*hand.Card, f HandCreationFunc) Hands {
	hands := map[int]*hand.Hand{}
	for seat, cards := range holeCards {
		hands[seat] = f(cards, board)
	}
	return hands
}

// WinningHands returns the highest ranking hands with the given sorting.
func (h Hands) WinningHands(sorting hand.Sorting) Hands {
	// copy all eligible hands (for Stud8 & Omaha8)
	handsCopy := Hands(map[int]*hand.Hand{})
	for seat, hand := range h {
		if hand != nil {
			handsCopy[seat] = hand
		}
	}

	if len(handsCopy) == 0 {
		return handsCopy
	}

	s := handsCopy.slice()
	s = hand.Sort(sorting, hand.DESC, s...)
	best := s[0]

	selected := map[int]*hand.Hand{}
	for seat, hand := range handsCopy {
		if best.CompareTo(hand) == 0 {
			selected[seat] = hand
		}
	}
	return Hands(selected)
}

func (h Hands) handsForSeats(seats []int) Hands {
	newHands := map[int]*hand.Hand{}
	for seat, hand := range h {
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

func (h Hands) slice() []*hand.Hand {
	s := []*hand.Hand{}
	for _, hand := range h {
		s = append(s, hand)
	}
	return s
}
