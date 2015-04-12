package game

import "github.com/loganjspears/joker/hand"

// HandCreationFunc represents the function signature for creating a hand.
type handCreationFunc func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand

// Hands represents a map of seats to hands
type hands map[int]*hand.Hand

// NewHands forms hands from seat's hole cards and the board.
func newHands(holeCards map[int][]*hand.Card, board []*hand.Card, f handCreationFunc) hands {
	hands := map[int]*hand.Hand{}
	for seat, cards := range holeCards {
		hands[seat] = f(cards, board)
	}
	return hands
}

// WinningHands returns the highest ranking hands with the given sorting.
func (h hands) winningHands(sorting hand.Sorting) hands {
	// copy all eligible hands (for Stud8 & Omaha8)
	handsCopy := hands(map[int]*hand.Hand{})
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
	return hands(selected)
}

func (h hands) formResults(solo Share, split Share) Results {
	share := solo
	if len(h) > 1 {
		share = split
	}

	results := map[int][]*Result{}
	for seat, hand := range h {
		result := &Result{
			Hand:  hand,
			Share: share,
		}
	}
	return Results(results)
}

func (h hands) handsForSeats(seats []int) hands {
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

func (h hands) slice() []*hand.Hand {
	s := []*hand.Hand{}
	for _, hand := range h {
		s = append(s, hand)
	}
	return s
}
