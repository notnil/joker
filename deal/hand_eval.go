package deal

import (
	"github.com/notnil/joker/hand"
	"github.com/notnil/joker/util"
)

type ranker struct {
	e handEvaluator
	s hand.Sorting
	o hand.Ordering
}

func (r *ranker) hands(holeCards map[int][]hand.Card, board []hand.Card) map[int]*hand.Hand {
	m := map[int]*hand.Hand{}
	for seat, hc := range holeCards {
		h := r.e.EvaluateHand(hc, board)
		m[seat] = h
	}
	return m
}

func (r *ranker) rank(hands map[int]*hand.Hand) [][]int {
	m := map[*hand.Hand]int{}
	order := []*hand.Hand{}
	for seat, h := range hands {
		if h != nil {
			m[h] = seat
			order = append(order, h)
		}
	}
	order = hand.Sort(r.s, r.o, order...)
	results := [][]int{}
	for i, h := range order {
		seat := m[h]
		if i == 0 || order[i-1].CompareTo(h) != 0 {
			results = append(results, []int{seat})
			continue
		}
		results[len(results)-1] = append(results[len(results)-1], seat)
	}
	return results
}

type handEvaluator interface {
	EvaluateHand(holeCards []hand.Card, board []hand.Card) *hand.Hand
}

type holdemHandEvaluator struct{}

func (holdemHandEvaluator) EvaluateHand(holeCards []hand.Card, board []hand.Card) *hand.Hand {
	cards := append(holeCards, board...)
	return hand.New(cards)
}

type omahaHiHandEvaluator struct{}

func (omahaHiHandEvaluator) EvaluateHand(holeCards []hand.Card, board []hand.Card) *hand.Hand {
	hands := []*hand.Hand{}
	for _, hc := range util.Combinations(4, 2) {
		for _, b := range util.Combinations(5, 3) {
			cards := []hand.Card{holeCards[hc[0]], holeCards[hc[1]], board[b[0]], board[b[1]], board[b[2]]}
			hands = append(hands, hand.New(cards))
		}
	}
	hands = hand.Sort(hand.SortingHigh, hand.DESC, hands...)
	return hands[0]
}

type omahaLowHandEvaluator struct{}

func (omahaLowHandEvaluator) EvaluateHand(holeCards []hand.Card, board []hand.Card) *hand.Hand {
	hands := []*hand.Hand{}
	for _, hc := range util.Combinations(4, 2) {
		for _, b := range util.Combinations(5, 3) {
			cards := []hand.Card{holeCards[hc[0]], holeCards[hc[1]], board[b[0]], board[b[1]], board[b[2]]}
			hands = append(hands, hand.New(cards, hand.AceToFiveLow))
		}
	}
	hands = hand.Sort(hand.SortingLow, hand.DESC, hands...)
	if isEightOrBetter(hands[0]) {
		return hands[0]
	}
	return nil
}

type razzHandEvaluator struct{}

func (razzHandEvaluator) EvaluateHand(holeCards []hand.Card, board []hand.Card) *hand.Hand {
	cards := append(holeCards, board...)
	return hand.New(cards, hand.AceToFiveLow)
}

type studLow8HandEvaluator struct{}

func (studLow8HandEvaluator) EvaluateHand(holeCards []hand.Card, board []hand.Card) *hand.Hand {
	cards := append(holeCards, board...)
	h := hand.New(cards, hand.AceToFiveLow)
	if isEightOrBetter(h) {
		return h
	}
	return nil
}

func isEightOrBetter(h *hand.Hand) bool {
	return h.CompareTo(eightOrBetter) <= 0
}

var (
	eightOrBetter = hand.New([]hand.Card{
		hand.EightSpades,
		hand.SevenSpades,
		hand.SixSpades,
		hand.FiveSpades,
		hand.FourSpades,
	}, hand.AceToFiveLow)
)
