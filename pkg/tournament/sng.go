// Package tournament implements single-table No-Limit Hold'em sit-and-gos
// and Independent Chip Model (ICM) equity.
package tournament

import (
	"errors"
	"slices"

	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/holdem"
)

var (
	ErrInvalidConfig    = errors.New("tournament: invalid config")
	ErrNotEnoughPlayers = errors.New("tournament: need at least two players")
	ErrTournamentOver   = errors.New("tournament: already finished")
	ErrHandNotFinished  = errors.New("tournament: hand is not finished")
)

// BlindLevel is one step of a hand-counted blind schedule.
type BlindLevel struct {
	SmallBlind int
	BigBlind   int
	Ante       int
	Hands      int // hands to stay at this level; 0 means stay forever
}

// Config describes a single-table tournament.
type Config struct {
	Seats         int
	StartingStack int
	Payouts       []int // index 0 is first place; should sum to the prize pool
	Blinds        []BlindLevel
}

// Entrant is a player to seat at the start of an SNG.
type Entrant struct {
	ID   string
	Seat int
}

// Finish is a player's finishing position and prize.
type Finish struct {
	Place int
	ID    string
	Seat  int
	Prize int
}

// SNG is a single-table sit-and-go.
type SNG struct {
	table        *holdem.Table
	cfg          Config
	level        int
	handsAtLevel int
	handsPlayed  int
	finishes     []Finish
	snapshot     map[int]int
	ids          map[int]string
	done         bool
}

// NewSNG seats players with equal starting stacks and points the button at
// the lowest live seat.
func NewSNG(cfg Config, players []Entrant, dealer hand.Dealer) (*SNG, error) {
	if cfg.Seats < 2 || cfg.Seats > 10 || cfg.StartingStack < 1 {
		return nil, ErrInvalidConfig
	}
	if len(cfg.Blinds) == 0 || cfg.Blinds[0].BigBlind < 1 {
		return nil, ErrInvalidConfig
	}
	if len(players) < 2 || len(players) > cfg.Seats {
		return nil, ErrNotEnoughPlayers
	}
	stakes := holdem.Stakes{
		SmallBlind: cfg.Blinds[0].SmallBlind,
		BigBlind:   cfg.Blinds[0].BigBlind,
		Ante:       cfg.Blinds[0].Ante,
	}
	tbl, err := holdem.NewTable(holdem.Config{Seats: cfg.Seats, Stakes: stakes}, dealer)
	if err != nil {
		return nil, err
	}
	ids := map[int]string{}
	for _, p := range players {
		if err := tbl.Sit(p.Seat, p.ID, cfg.StartingStack); err != nil {
			return nil, err
		}
		ids[p.Seat] = p.ID
	}
	s := &SNG{
		table: tbl,
		cfg:   cfg,
		ids:   ids,
	}
	_ = tbl.SetButton(tbl.LiveSeats()[0])
	return s, nil
}

// Table is the underlying Hold'em table.
func (s *SNG) Table() *holdem.Table { return s.table }

// Level is the current blind level.
func (s *SNG) Level() BlindLevel { return s.cfg.Blinds[s.level] }

// HandsPlayed is the number of completed hands.
func (s *SNG) HandsPlayed() int { return s.handsPlayed }

// Done is true when one player remains.
func (s *SNG) Done() bool { return s.done }

// Standings are finishes from first place to last, once the tournament is over.
// During play it contains busted players only (worst-first as they bust).
func (s *SNG) Standings() []Finish {
	out := slices.Clone(s.finishes)
	slices.SortFunc(out, func(a, b Finish) int { return a.Place - b.Place })
	return out
}

// StartHand deals the next tournament hand at the current blinds.
func (s *SNG) StartHand() (*holdem.Hand, error) {
	if s.done {
		return nil, ErrTournamentOver
	}
	if len(s.table.LiveSeats()) < 2 {
		s.finishWinner()
		return nil, ErrTournamentOver
	}
	lv := s.Level()
	if err := s.table.SetStakes(holdem.Stakes{SmallBlind: lv.SmallBlind, BigBlind: lv.BigBlind, Ante: lv.Ante}); err != nil {
		return nil, err
	}
	s.snapshot = map[int]int{}
	for _, seat := range s.table.LiveSeats() {
		s.snapshot[seat] = s.table.Player(seat).Chips
	}
	return s.table.StartHand()
}

// FinishHand applies results, eliminates busted players, and may end the SNG.
func (s *SNG) FinishHand(h *holdem.Hand) error {
	if s.done {
		return ErrTournamentOver
	}
	if h == nil || !h.Done() {
		return ErrHandNotFinished
	}
	button := h.Button()
	if err := s.table.ApplyResults(h); err != nil {
		return err
	}
	s.handsPlayed++
	s.handsAtLevel++
	s.eliminate(button)
	if len(s.table.LiveSeats()) <= 1 {
		s.finishWinner()
		return nil
	}
	s.maybeAdvanceBlinds()
	return nil
}

func (s *SNG) eliminate(button int) {
	var busted []int
	for seat := range s.snapshot {
		p := s.table.Player(seat)
		if p != nil && p.Chips == 0 {
			busted = append(busted, seat)
		}
	}
	if len(busted) == 0 {
		return
	}
	size := s.cfg.Seats
	slices.SortFunc(busted, func(a, b int) int {
		if s.snapshot[a] != s.snapshot[b] {
			return s.snapshot[a] - s.snapshot[b] // fewer chips = worse finish
		}
		return distLeft(button, b, size) - distLeft(button, a, size)
	})
	liveAfter := len(s.table.LiveSeats())
	for i, seat := range busted {
		place := liveAfter + len(busted) - i
		s.record(place, seat)
		_ = s.table.StandUp(seat)
	}
}

func (s *SNG) finishWinner() {
	if s.done {
		return
	}
	s.done = true
	live := s.table.LiveSeats()
	if len(live) == 1 {
		s.record(1, live[0])
	}
}

func (s *SNG) record(place, seat int) {
	prize := 0
	if place >= 1 && place-1 < len(s.cfg.Payouts) {
		prize = s.cfg.Payouts[place-1]
	}
	s.finishes = append(s.finishes, Finish{
		Place: place,
		ID:    s.ids[seat],
		Seat:  seat,
		Prize: prize,
	})
}

func (s *SNG) maybeAdvanceBlinds() {
	lv := s.Level()
	if lv.Hands <= 0 || s.handsAtLevel < lv.Hands {
		return
	}
	if s.level+1 >= len(s.cfg.Blinds) {
		return
	}
	s.level++
	s.handsAtLevel = 0
}

func distLeft(button, seat, tableSize int) int {
	return (seat - button - 1 + tableSize) % tableSize
}
