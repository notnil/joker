package game

import (
	"github.com/notnil/joker/hand"
	"github.com/notnil/joker/util"
)

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
