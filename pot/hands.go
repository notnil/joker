package pot

import "github.com/SyntropyDev/joker/hand"

// type handCreationFunc func(holeCards []*hand.Card, board []*hand.Card) *hand.Hand
//

type hands map[int]*hand.Hand

func (hnds hands) handsForSeats(seats []int) hands {
	newHands := map[int]*hand.Hand{}
	for seat, hand := range hnds {
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

func (hnds hands) winningHands(sorting hand.Sorting) hands {
	// copy all eligible hands (for Stud8 & Omaha8)
	handsCopy := hands(map[int]*hand.Hand{})
	for seat, hand := range hnds {
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

func (hnds hands) slice() []*hand.Hand {
	s := []*hand.Hand{}
	for _, h := range hnds {
		s = append(s, h)
	}
	return s
}
