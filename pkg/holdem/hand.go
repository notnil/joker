package holdem

import (
	"slices"

	"github.com/notnil/joker/pkg/hand"
)

// Street is a betting round.
type Street int

const (
	Preflop Street = iota
	Flop
	Turn
	River
)

func (s Street) String() string {
	switch s {
	case Preflop:
		return "Preflop"
	case Flop:
		return "Flop"
	case Turn:
		return "Turn"
	case River:
		return "River"
	default:
		return "Street(?)"
	}
}

type playerState struct {
	id          string
	seat        int
	chips       int
	hole        []hand.Card
	folded      bool
	allIn       bool
	streetPut   int
	totalPut    int
	needsAction bool
	canRaise    bool
}

// Result is a player's outcome for a finished hand.
type Result struct {
	Seat   int
	Award  int
	Hand   *hand.Hand // nil if the player folded or won without showdown
	Folded bool
}

// Hand is one deal of No-Limit Hold'em.
type Hand struct {
	table         *Table
	players       map[int]*playerState
	pot           *Pot
	deck          *hand.Deck
	board         []hand.Card
	button        int
	sb            int
	bb            int
	toAct         int
	street        Street
	currentBet    int
	lastFullRaise int
	betOpened     bool
	done          bool
	results       []Result
}

func newHand(t *Table) *Hand {
	players := map[int]*playerState{}
	for _, seat := range t.LiveSeats() {
		p := t.seats[seat]
		players[seat] = &playerState{
			id:    p.ID,
			seat:  seat,
			chips: p.Chips,
		}
	}
	return &Hand{
		table:   t,
		players: players,
		pot:     NewPot(),
		deck:    t.dealer.Deck(),
		button:  t.button,
		toAct:   -1,
	}
}

func (h *Hand) setupPreflop() {
	h.street = Preflop
	live := h.liveSeats()
	if len(live) == 2 {
		h.sb = h.button
		h.bb = h.nextInHand(h.button)
	} else {
		h.sb = h.nextInHand(h.button)
		h.bb = h.nextInHand(h.sb)
	}

	order := h.dealOrder()
	for _, seat := range order {
		h.players[seat].hole = h.deck.PopMulti(2)
	}

	stakes := h.table.config.Stakes
	if stakes.Ante > 0 {
		for _, seat := range order {
			h.take(seat, stakes.Ante, false)
		}
	}
	h.take(h.sb, stakes.SmallBlind, true)
	h.take(h.bb, stakes.BigBlind, true)

	h.betOpened = true
	h.currentBet = 0
	for _, p := range h.players {
		if p.streetPut > h.currentBet {
			h.currentBet = p.streetPut
		}
	}
	h.lastFullRaise = stakes.BigBlind
	for _, p := range h.players {
		if p.folded || p.allIn {
			p.needsAction = false
			continue
		}
		p.needsAction = true
		p.canRaise = true
	}
	h.toAct = h.nextInHand(h.bb)
}

func (h *Hand) dealOrder() []int {
	var order []int
	start := h.sb
	seat := start
	for {
		order = append(order, seat)
		seat = h.nextInHand(seat)
		if seat == start {
			break
		}
	}
	return order
}

// Street is the current betting round.
func (h *Hand) Street() Street { return h.street }

// Board is the community cards.
func (h *Hand) Board() []hand.Card {
	return slices.Clone(h.board)
}

// Hole returns the hole cards for seat.
func (h *Hand) Hole(seat int) []hand.Card {
	p := h.players[seat]
	if p == nil {
		return nil
	}
	return slices.Clone(p.hole)
}

// ToAct is the seat that must act, or -1 if the hand is over or running out.
func (h *Hand) ToAct() int { return h.toAct }

// Button is the dealer-button seat for this hand.
func (h *Hand) Button() int { return h.button }

// SmallBlindSeat is the seat that posted the small blind.
func (h *Hand) SmallBlindSeat() int { return h.sb }

// BigBlindSeat is the seat that posted the big blind.
func (h *Hand) BigBlindSeat() int { return h.bb }

// Stack is the remaining chips for seat in this hand (not including pot awards).
func (h *Hand) Stack(seat int) int {
	p := h.players[seat]
	if p == nil {
		return 0
	}
	return p.chips
}

// Folded reports whether seat has folded.
func (h *Hand) Folded(seat int) bool {
	p := h.players[seat]
	return p != nil && p.folded
}

// IsAllIn reports whether seat is all-in.
func (h *Hand) IsAllIn(seat int) bool {
	p := h.players[seat]
	return p != nil && p.allIn
}

// PotTotal is chips in the pot.
func (h *Hand) PotTotal() int { return h.pot.Total() }

// Pots returns main and side pots. Eligible seats are contesting that pot.
func (h *Hand) Pots() []SidePot { return h.pot.SidePots() }

// Done is true after showdown or a fold-win.
func (h *Hand) Done() bool { return h.done }

// Results is populated when the hand is done.
func (h *Hand) Results() []Result { return slices.Clone(h.results) }

// Awards maps seat to chips won from the pot (including returned uncalled bets).
func (h *Hand) Awards() map[int]int {
	out := map[int]int{}
	for _, r := range h.results {
		out[r.Seat] = r.Award
	}
	return out
}

// CurrentBet is the largest contribution this street.
func (h *Hand) CurrentBet() int { return h.currentBet }

// StreetPut is chips the seat has put in this street.
func (h *Hand) StreetPut(seat int) int {
	p := h.players[seat]
	if p == nil {
		return 0
	}
	return p.streetPut
}

// CanRaise reports whether the current actor may raise.
func (h *Hand) CanRaise() bool {
	p := h.actor()
	return p != nil && p.canRaise
}

func (h *Hand) CallAmount() int {
	p := h.actor()
	if p == nil {
		return 0
	}
	owe := h.currentBet - p.streetPut
	if owe < 0 {
		return 0
	}
	return min(owe, p.chips)
}

// MinBet is the minimum opening bet on this street (the big blind).
func (h *Hand) MinBet() int {
	return h.table.config.Stakes.BigBlind
}

// MinRaiseTo is the minimum legal raise-to amount, or 0 if a raise is illegal.
func (h *Hand) MinRaiseTo() int {
	return h.currentBet + h.lastFullRaise
}

// LegalActions lists action types the current player may take.
func (h *Hand) LegalActions() []ActionType {
	if h.done {
		return nil
	}
	p := h.actor()
	if p == nil {
		return nil
	}
	owe := h.currentBet - p.streetPut
	acts := []ActionType{Fold}
	if owe <= 0 {
		acts = append(acts, Check)
		if p.chips > 0 {
			if p.canRaise && p.chips >= h.MinBet() {
				acts = append(acts, Bet)
			}
			acts = append(acts, AllIn)
		}
		return uniqActions(acts)
	}
	if p.chips > owe {
		acts = append(acts, Call)
		if p.canRaise {
			acts = append(acts, AllIn)
			if p.streetPut+p.chips >= h.MinRaiseTo() {
				acts = append(acts, Raise)
			}
		}
		return uniqActions(acts)
	}
	acts = append(acts, Call, AllIn)
	return uniqActions(acts)
}

// Act applies an action for the player to act.
func (h *Hand) Act(a Action) error {
	if h.done {
		return ErrInvalidAction
	}
	p := h.actor()
	if p == nil {
		return ErrInvalidAction
	}
	if !includesType(h.LegalActions(), a.Type) {
		return ErrInvalidAction
	}
	owe := h.currentBet - p.streetPut
	if owe < 0 {
		owe = 0
	}

	switch a.Type {
	case Fold:
		p.folded = true
		p.needsAction = false
		h.pot.Fold(p.seat)
	case Check:
		if owe != 0 {
			return ErrInvalidAction
		}
		p.needsAction = false
	case Call:
		h.take(p.seat, owe, true)
		p.needsAction = false
	case Bet:
		if owe != 0 || a.Chips < h.MinBet() || a.Chips > p.chips {
			return ErrInvalidAmount
		}
		h.take(p.seat, a.Chips, true)
		h.onAggressive(p.seat, a.Chips >= h.MinBet())
		p.needsAction = false
	case Raise:
		raiseTo := a.Chips
		if raiseTo < h.MinRaiseTo() || raiseTo <= h.currentBet {
			return ErrInvalidAmount
		}
		need := raiseTo - p.streetPut
		if need <= 0 || need > p.chips {
			return ErrInvalidAmount
		}
		raiseSize := raiseTo - h.currentBet
		h.take(p.seat, need, true)
		h.onAggressive(p.seat, raiseSize >= h.lastFullRaise)
		p.needsAction = false
	case AllIn:
		put := p.chips
		newBet := p.streetPut + put
		h.take(p.seat, put, true)
		if newBet > h.currentBet {
			var full bool
			if !h.betOpened {
				full = newBet >= h.MinBet()
			} else {
				full = (newBet - h.currentBet) >= h.lastFullRaise
			}
			h.onAggressive(p.seat, full)
		}
		p.needsAction = false
	default:
		return ErrInvalidAction
	}
	h.continueHand()
	return nil
}

func (h *Hand) onAggressive(seat int, full bool) {
	p := h.players[seat]
	newBet := p.streetPut
	raiseSize := newBet - h.currentBet
	if !h.betOpened {
		h.betOpened = true
		if full {
			h.lastFullRaise = newBet
		} else {
			h.lastFullRaise = h.MinBet()
		}
	} else if full && raiseSize > 0 {
		h.lastFullRaise = raiseSize
	}
	h.currentBet = newBet
	for _, o := range h.players {
		if o == nil || o.seat == seat || o.folded || o.allIn {
			continue
		}
		if full {
			o.needsAction = true
			o.canRaise = true
			continue
		}
		if o.streetPut < h.currentBet {
			if !o.needsAction {
				o.canRaise = false
			}
			o.needsAction = true
		}
	}
}

func (h *Hand) continueHand() {
	if h.done {
		return
	}
	if n := h.contesting(); len(n) <= 1 {
		h.foldWin()
		return
	}
	if h.someoneNeedsAction() {
		p := h.players[h.toAct]
		if p == nil || !p.needsAction || p.folded || p.allIn {
			from := h.toAct
			if from < 0 {
				from = h.button
			}
			h.toAct = h.nextNeedingAction(from)
		}
		return
	}
	if h.street == River {
		h.showdown()
		return
	}
	h.advanceStreet()
	h.continueHand()
}

func (h *Hand) advanceStreet() {
	h.street++
	switch h.street {
	case Flop:
		h.board = h.deck.PopMulti(3)
	case Turn, River:
		h.board = append(h.board, h.deck.Pop())
	}
	h.currentBet = 0
	h.lastFullRaise = h.MinBet()
	h.betOpened = false
	canAct := 0
	for _, p := range h.players {
		if !p.folded && !p.allIn {
			canAct++
		}
	}
	betting := canAct >= 2
	for _, p := range h.players {
		p.streetPut = 0
		if !betting || p.folded || p.allIn {
			p.needsAction = false
			p.canRaise = false
			continue
		}
		p.needsAction = true
		p.canRaise = true
	}
	if betting {
		h.toAct = h.nextNeedingAction(h.button)
	} else {
		h.toAct = -1
	}
}

func (h *Hand) take(seat, chips int, street bool) {
	p := h.players[seat]
	if chips <= 0 || p.chips <= 0 {
		return
	}
	amount := min(chips, p.chips)
	p.chips -= amount
	p.totalPut += amount
	if street {
		p.streetPut += amount
	}
	h.pot.Add(seat, amount)
	if p.chips == 0 {
		p.allIn = true
		p.needsAction = false
	}
}

func (h *Hand) contribute(seat, chips int) {
	h.take(seat, chips, true)
}

func (h *Hand) actor() *playerState {
	if h.toAct < 0 {
		return nil
	}
	return h.players[h.toAct]
}

func (h *Hand) someoneNeedsAction() bool {
	for _, p := range h.players {
		if p.needsAction && !p.folded && !p.allIn {
			return true
		}
	}
	return false
}

func (h *Hand) nextNeedingAction(from int) int {
	n := h.table.config.Seats
	for i := 1; i <= n; i++ {
		s := (from + i) % n
		p := h.players[s]
		if p != nil && p.needsAction && !p.folded && !p.allIn {
			return s
		}
	}
	return -1
}

func (h *Hand) nextInHand(from int) int {
	return h.table.nextOccupiedIn(from, h.players)
}

func (h *Hand) liveSeats() []int {
	var seats []int
	for seat := range h.table.config.Seats {
		if h.players[seat] != nil {
			seats = append(seats, seat)
		}
	}
	return seats
}

func (h *Hand) contesting() []*playerState {
	var out []*playerState
	for seat := range h.table.config.Seats {
		p := h.players[seat]
		if p != nil && !p.folded {
			out = append(out, p)
		}
	}
	return out
}

func uniqActions(in []ActionType) []ActionType {
	seen := map[ActionType]bool{}
	var out []ActionType
	for _, a := range in {
		if seen[a] {
			continue
		}
		seen[a] = true
		out = append(out, a)
	}
	return out
}
