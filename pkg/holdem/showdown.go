package holdem

import (
	"slices"

	"github.com/notnil/joker/pkg/hand"
)

func (h *Hand) foldWin() {
	h.done = true
	h.toAct = -1
	var winner *playerState
	for _, p := range h.contesting() {
		winner = p
		break
	}
	h.results = h.baseResults()
	if winner == nil {
		return
	}
	for i := range h.results {
		if h.results[i].Seat == winner.seat {
			h.results[i].Award = h.pot.Total()
			h.results[i].Hand = nil
		}
	}
}

func (h *Hand) showdown() {
	h.done = true
	h.toAct = -1
	h.results = h.baseResults()
	award := map[int]int{}
	shown := map[int]*hand.Hand{}

	for _, p := range h.contesting() {
		cards := append(slices.Clone(p.hole), h.board...)
		shown[p.seat] = hand.New(cards)
	}

	size := h.table.config.Seats
	for _, sp := range h.pot.SidePots() {
		if len(sp.Eligible) == 0 || sp.Amount == 0 {
			continue
		}
		winners := bestSeats(sp.Eligible, shown)
		order := oddChipOrder(winners, h.button, size)
		for seat, chips := range splitAmount(sp.Amount, len(order), order) {
			award[seat] += chips
		}
	}

	for i := range h.results {
		seat := h.results[i].Seat
		h.results[i].Award = award[seat]
		if !h.results[i].Folded {
			h.results[i].Hand = shown[seat]
		}
	}
}

func (h *Hand) baseResults() []Result {
	var out []Result
	for seat := range h.table.config.Seats {
		p := h.players[seat]
		if p == nil {
			continue
		}
		out = append(out, Result{
			Seat:   seat,
			Folded: p.folded,
		})
	}
	return out
}

func bestSeats(eligible []int, hands map[int]*hand.Hand) []int {
	var winners []int
	for _, seat := range eligible {
		hs := hands[seat]
		if hs == nil {
			continue
		}
		if len(winners) == 0 {
			winners = []int{seat}
			continue
		}
		cmp := hs.CompareTo(hands[winners[0]])
		if cmp > 0 {
			winners = []int{seat}
		} else if cmp == 0 {
			winners = append(winners, seat)
		}
	}
	return winners
}
