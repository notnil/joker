package table

import (
	"errors"
	"sort"

	"github.com/notnil/joker/hand"
)

var (
	ErrInvalidAction    = errors.New("attempted invalid action")
	ErrInvalidBetAmount = errors.New("invalid bet amount")
)

type Round int

const (
	PreFlop Round = iota
	Flop
	Turn
	River
)

type Action struct {
	Type  ActionType
	Chips int
}

type ActionType int

const (
	Fold ActionType = iota
	Check
	Call
	Bet
	Raise
	AllIn
)

var (
	actionTypeNames = []string{"Fold", "Check", "Call", "Bet", "Raise", "AllIn"}
)

func (at ActionType) String() string {
	return actionTypeNames[at]
}

type Hand struct {
	Table   *Table
	Pot     *Pot
	Deck    *hand.Deck
	Seats   map[int]*PlayerInHand
	Board   []hand.Card
	Active  int
	Round   Round
	Results map[int][]HandResult
}

type PlayerInHand struct {
	ID     string
	Seat   int
	Chips  int
	Acted  bool
	Folded bool
	AllIn  bool
	Cards  []hand.Card
}

func (h *Hand) Fold() error {
	return h.Act(Action{Type: Fold})
}

func (h *Hand) Check() error {
	return h.Act(Action{Type: Check})
}

func (h *Hand) Call() error {
	return h.Act(Action{Type: Call})
}

func (h *Hand) Bet(chips int) error {
	return h.Act(Action{Type: Bet, Chips: chips})
}

func (h *Hand) Raise(chips int) error {
	return h.Act(Action{Type: Raise, Chips: chips})
}

func (h *Hand) AllIn() error {
	return h.Act(Action{Type: AllIn})
}

func (h *Hand) Act(a Action) error {
	if h.Results != nil {
		return ErrInvalidAction
	}
	if !includes(h.LegalActions(), a.Type) {
		return ErrInvalidAction
	}
	// TODO enforce limits, min bets
	player := h.ActivePlayer()
	owe := h.Pot.Owe(h.Active)
	switch a.Type {
	case Fold:
		player.Folded = true
		h.Pot.Remove(player.Seat)
	case Check:
	case Call:
		h.contribute(player, owe)
	case Bet, Raise, AllIn:
		if (a.Chips < h.Table.config.Stakes.BigBlind) && a.Type != AllIn {
			return ErrInvalidBetAmount
		}
		h.contribute(player, owe)
		h.contribute(player, a.Chips)
		h.resetAction()
	}
	player.Acted = true
	h.update()
	return nil
}

func (h *Hand) LegalActions() []ActionType {
	owe := h.Pot.Owe(h.Active)
	if owe == 0 {
		return []ActionType{Fold, Check, Bet, AllIn}
	}
	if owe > h.ActivePlayer().Chips {
		return []ActionType{Fold, Call}
	}
	return []ActionType{Fold, Call, Raise, AllIn}
}

func (h *Hand) ActivePlayer() *PlayerInHand {
	return h.Seats[h.Active]
}

func (h *Hand) update() {
	seat := h.nextToAct()
	if seat != -1 {
		h.Active = seat
		return
	}
	if len(h.contesting()) == 1 || h.Round == River {
		h.calcResults()
		return
	}
	h.Round++
	h.setupRound()
}

func (h *Hand) setupRound() {
	h.resetAction()
	switch h.Round {
	case PreFlop:
		sb := h.Table.Next(h.Table.button)
		bb := h.Table.Next(sb)
		if h.Table.PlayerCount() == 2 {
			sb = h.Table.button
			bb = h.Table.Next(h.Table.button)
		}
		h.Deck = h.Table.dealer.Deck()
		for _, seat := range h.orderedSeats() {
			player := h.Seats[seat]
			player.Cards = h.Deck.PopMulti(2)
			h.contribute(player, h.Table.config.Stakes.Ante)
		}
		h.contribute(h.Seats[sb], h.Table.config.Stakes.SmallBlind)
		h.contribute(h.Seats[bb], h.Table.config.Stakes.BigBlind)
		h.Active = h.Table.Next(bb)
		// TODO bug w/ big blind having to go all and cost < bb
	case Flop:
		h.Board = h.Deck.PopMulti(3)
		h.Active = h.Table.Next(h.Table.button)
	case Turn, River:
		h.Board = append(h.Board, h.Deck.Pop())
		h.Active = h.Table.Next(h.Table.button)
	}
}

type PotShare int

const (
	Won PotShare = iota
	Split
)

type HandResult struct {
	Hand     *hand.Hand
	PotShare PotShare
	Chips    int
}

func (h *Hand) calcResults() {
	if len(h.contesting()) == 1 {
		seat := h.contesting()[0].Seat
		h.Results = map[int][]HandResult{}
		h.Results[seat] = []HandResult{{
			Hand:     nil,
			PotShare: Won,
			Chips:    h.Pot.Total(),
		}}
		return
	}
	hands := map[int]*hand.Hand{}
	for seat, player := range h.Seats {
		// TODO omaha
		hands[seat] = hand.New(append(player.Cards, h.Board...))
	}
	results := map[int][]HandResult{}
	for _, pot := range h.Pot.Split() {
		// sort by best hand first
		elegible := pot.Eligible()
		sort.Slice(elegible, func(i, j int) bool {
			iHand := hands[elegible[i]]
			jHand := hands[elegible[j]]
			return iHand.CompareTo(jHand) > 0
		})
		// select winners who split pot if more than one
		winners := []int{}
		h1 := hands[elegible[0]]
		for seat := range elegible {
			h2 := hands[seat]
			if h1.CompareTo(h2) != 0 {
				break
			}
			winners = append(winners, seat)
		}
		// sort closest to the button for spare chips in split pot
		sort.Slice(winners, func(i, j int) bool {
			iDist := h.distanceFromButton(winners[i])
			jDist := h.distanceFromButton(winners[j])
			return iDist < jDist
		})
		// payout chips
		for i, seat := range winners {
			chips := pot.Total() / len(winners)
			if (pot.Total() % len(winners)) > i {
				chips++
			}
			potshare := Won
			if len(winners) > 1 {
				potshare = Split
			}
			result := HandResult{
				Hand:     hands[seat],
				PotShare: potshare,
				Chips:    chips,
			}
			results[seat] = append(results[seat], result)
		}
	}
	h.Results = results
}

func (h *Hand) contribute(p *PlayerInHand, chips int) {
	amount := chips
	if p.Chips <= amount {
		amount = p.Chips
		p.AllIn = true
	}
	p.Chips -= amount
	h.Pot.Add(p.Seat, chips)
}

func (h *Hand) resetAction() {
	for _, seat := range h.Seats {
		if seat != nil {
			seat.Acted = false
		}
	}
}

func (h *Hand) nextToAct() int {
	seat := h.Active
	for i := 0; i < len(h.Seats); i++ {
		seat = h.Table.Next(seat)
		player := h.Seats[seat]
		if !player.Acted && !player.AllIn && !player.Folded {
			return player.Seat
		}
	}
	return -1
}

func (h *Hand) contesting() []*PlayerInHand {
	contesting := []*PlayerInHand{}
	for _, seat := range h.Seats {
		if !seat.Folded {
			contesting = append(contesting, seat)
		}
	}
	return contesting
}

func (h *Hand) distanceFromButton(seat int) int {
	cur := h.Table.button
	dist := 0
	for {
		cur = h.Table.Next(cur)
		dist++
		if cur == seat {
			return dist
		}
	}
}

func (h *Hand) orderedSeats() []int {
	seats := []int{}
	for seat := range h.Seats {
		seats = append(seats, seat)
	}
	sort.Slice(seats, func(i, j int) bool {
		iDist := h.distanceFromButton(seats[i])
		jDist := h.distanceFromButton(seats[j])
		return iDist < jDist
	})
	return seats
}

func includes(actions []ActionType, include ...ActionType) bool {
	for _, a1 := range include {
		found := false
		for _, a2 := range actions {
			found = found || a1 == a2
		}
		if !found {
			return false
		}
	}
	return true
}
