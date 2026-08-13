package holdem

import (
	"errors"
	"fmt"

	"github.com/notnil/joker/pkg/hand"
)

var (
	ErrInvalidSeatCount = errors.New("holdem: tables must have between 2 and 10 seats")
	ErrInvalidStakes    = errors.New("holdem: big blind must be positive")
	ErrInvalidSeat      = errors.New("holdem: invalid seat")
	ErrSeatOccupied     = errors.New("holdem: seat occupied")
	ErrSeatEmpty        = errors.New("holdem: seat empty")
	ErrDuplicatePlayer  = errors.New("holdem: player already seated")
	ErrNotEnoughPlayers = errors.New("holdem: need at least two players with chips")
	ErrHandInProgress   = errors.New("holdem: hand already in progress")
	ErrNoHand           = errors.New("holdem: no hand in progress")
	ErrInvalidAction    = errors.New("holdem: illegal action")
	ErrInvalidAmount    = errors.New("holdem: invalid chip amount")
	ErrNilDealer        = errors.New("holdem: dealer is required")
)

// Stakes are the forced bets for a hand, in integer chips.
type Stakes struct {
	SmallBlind int
	BigBlind   int
	Ante       int
}

// Config configures a table.
type Config struct {
	Seats  int
	Stakes Stakes
}

// Player is a seated player.
type Player struct {
	ID    string
	Chips int
}

// Table is a No-Limit Hold'em cash or tournament table.
type Table struct {
	seats  []*Player
	config Config
	dealer hand.Dealer
	button int
}

// NewTable creates an empty table. dealer must be non-nil.
func NewTable(cfg Config, dealer hand.Dealer) (*Table, error) {
	if cfg.Seats < 2 || cfg.Seats > 10 {
		return nil, ErrInvalidSeatCount
	}
	if cfg.Stakes.BigBlind < 1 {
		return nil, ErrInvalidStakes
	}
	if cfg.Stakes.SmallBlind < 0 || cfg.Stakes.Ante < 0 {
		return nil, ErrInvalidStakes
	}
	if dealer == nil {
		return nil, ErrNilDealer
	}
	return &Table{
		seats:  make([]*Player, cfg.Seats),
		config: cfg,
		dealer: dealer,
		button: 0,
	}, nil
}

// Sit seats a player. chips must be positive.
func (t *Table) Sit(seat int, id string, chips int) error {
	if seat < 0 || seat >= t.config.Seats {
		return ErrInvalidSeat
	}
	if t.seats[seat] != nil {
		return ErrSeatOccupied
	}
	if chips <= 0 {
		return ErrInvalidAmount
	}
	for _, p := range t.seats {
		if p != nil && p.ID == id {
			return ErrDuplicatePlayer
		}
	}
	t.seats[seat] = &Player{ID: id, Chips: chips}
	return nil
}

// StandUp removes the player at seat.
func (t *Table) StandUp(seat int) error {
	if seat < 0 || seat >= t.config.Seats {
		return ErrInvalidSeat
	}
	if t.seats[seat] == nil {
		return ErrSeatEmpty
	}
	t.seats[seat] = nil
	return nil
}

// Player returns a copy of the seated player, or nil.
func (t *Table) Player(seat int) *Player {
	if seat < 0 || seat >= t.config.Seats || t.seats[seat] == nil {
		return nil
	}
	cp := *t.seats[seat]
	return &cp
}

// Button is the dealer-button seat index.
func (t *Table) Button() int { return t.button }

// SetButton points the button at a live seat.
func (t *Table) SetButton(seat int) error {
	if !t.isLive(seat) {
		return ErrInvalidSeat
	}
	t.button = seat
	return nil
}

// SetStakes updates blinds/antes between hands.
func (t *Table) SetStakes(s Stakes) error {
	if s.BigBlind < 1 || s.SmallBlind < 0 || s.Ante < 0 {
		return ErrInvalidStakes
	}
	t.config.Stakes = s
	return nil
}

// Stakes returns the current stakes.
func (t *Table) Stakes() Stakes { return t.config.Stakes }

// Seats is the table size.
func (t *Table) Seats() int { return t.config.Seats }

// LiveSeats returns occupied seats with chips, in seat order.
func (t *Table) LiveSeats() []int {
	var seats []int
	for i, p := range t.seats {
		if p != nil && p.Chips > 0 {
			seats = append(seats, i)
		}
	}
	return seats
}

// StartHand deals a new hand. At least two live players are required.
func (t *Table) StartHand() (*Hand, error) {
	t.ensureButton()
	live := t.LiveSeats()
	if len(live) < 2 {
		return nil, ErrNotEnoughPlayers
	}
	h := newHand(t)
	h.setupPreflop()
	h.continueHand()
	return h, nil
}

// ApplyResults writes stack changes back to the table and advances the button.
func (t *Table) ApplyResults(h *Hand) error {
	if h == nil || !h.Done() {
		return ErrNoHand
	}
	awards := h.Awards()
	for seat, p := range h.players {
		if p == nil {
			continue
		}
		tp := t.seats[seat]
		if tp == nil {
			continue
		}
		tp.Chips = p.chips + awards[seat]
	}
	t.button = t.nextLive(t.button)
	return nil
}

func (t *Table) ensureButton() {
	if t.isLive(t.button) {
		return
	}
	if s := t.nextLive(t.button); s >= 0 {
		t.button = s
	}
}

func (t *Table) isLive(seat int) bool {
	if seat < 0 || seat >= t.config.Seats {
		return false
	}
	p := t.seats[seat]
	return p != nil && p.Chips > 0
}

func (t *Table) nextLive(from int) int {
	n := t.config.Seats
	for i := 1; i <= n; i++ {
		s := (from + i) % n
		if t.isLive(s) {
			return s
		}
	}
	return -1
}

func (t *Table) nextOccupiedIn(from int, in map[int]*playerState) int {
	n := t.config.Seats
	for i := 1; i <= n; i++ {
		s := (from + i) % n
		if in[s] != nil {
			return s
		}
	}
	return -1
}

func (t *Table) String() string {
	return fmt.Sprintf("table button=%d stakes=%d/%d", t.button, t.config.Stakes.SmallBlind, t.config.Stakes.BigBlind)
}
