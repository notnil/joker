package table

// Table provides poker-table-like operations: seating and action flow.
type Table interface {
    // Seating
    SeatAt(seat SeatIndex, player Player) error
    AutoSeat(player Player) (SeatIndex, error)
    Leave(seat SeatIndex) error
    IsOccupied(seat SeatIndex) (bool, error)
    FindPlayer(playerID PlayerID) (SeatIndex, bool)
    NumSeated() int
    MaxSeats() int

    // Action Flow
    StartHand() error
    EndHand()
    Dealer() SeatIndex
    ToAct() SeatIndex
    NextToAct() (SeatIndex, error)
    AdvanceAction() (SeatIndex, error)
    AdvanceDealer() error

    // Snapshot state for external use
    Snapshot() State
}

// New returns a Table with the given configuration.
func New(cfg Config) Table {
    if cfg.MaxSeats <= 0 {
        cfg.MaxSeats = 9
    }
    s := make([]Seat, cfg.MaxSeats)
    for i := range s {
        s[i] = Seat{Index: SeatIndex(i), Status: SeatEmpty}
    }
    return &tableImpl{
        cfg:   cfg,
        state: State{Seats: s, Action: ActionPosition{Dealer: NoSeat, ToAct: NoSeat, HandActive: false}},
    }
}

type tableImpl struct {
    cfg   Config
    state State
}

func (t *tableImpl) MaxSeats() int { return t.cfg.MaxSeats }
func (t *tableImpl) NumSeated() int { return t.state.Seated }

func (t *tableImpl) Snapshot() State {
    // Return a shallow copy to prevent external mutation.
    s := make([]Seat, len(t.state.Seats))
    copy(s, t.state.Seats)
    return State{Seats: s, Action: t.state.Action, Seated: t.state.Seated}
}

func (t *tableImpl) IsOccupied(seat SeatIndex) (bool, error) {
    if !t.validSeat(seat) {
        return false, ErrInvalidSeat
    }
    return t.state.Seats[seat].Status == SeatOccupied, nil
}

func (t *tableImpl) SeatAt(seat SeatIndex, player Player) error {
    if !t.validSeat(seat) {
        return ErrInvalidSeat
    }
    if _, ok := t.FindPlayer(player.ID); ok {
        return ErrPlayerSeated
    }
    if t.state.Seats[seat].Status == SeatOccupied {
        return ErrSeatOccupied
    }
    t.state.Seats[seat].Status = SeatOccupied
    t.state.Seats[seat].Player = &player
    t.state.Seated++
    // Initialize dealer to first seated player if none.
    if t.state.Action.Dealer == NoSeat {
        t.state.Action.Dealer = seat
    }
    return nil
}

func (t *tableImpl) AutoSeat(player Player) (SeatIndex, error) {
    if _, ok := t.FindPlayer(player.ID); ok {
        return NoSeat, ErrPlayerSeated
    }
    if t.state.Seated >= t.cfg.MaxSeats {
        return NoSeat, ErrTableFull
    }
    for i := 0; i < t.cfg.MaxSeats; i++ {
        if t.state.Seats[i].Status == SeatEmpty {
            return SeatIndex(i), t.SeatAt(SeatIndex(i), player)
        }
    }
    return NoSeat, ErrTableFull
}

func (t *tableImpl) Leave(seat SeatIndex) error {
    if !t.validSeat(seat) {
        return ErrInvalidSeat
    }
    if t.state.Seats[seat].Status == SeatEmpty {
        return ErrSeatEmpty
    }
    t.state.Seats[seat].Status = SeatEmpty
    t.state.Seats[seat].Player = nil
    t.state.Seated--
    // If dealer or to-act left, adjust.
    if t.state.Action.Dealer == seat {
        t.reassignDealer()
    }
    if t.state.Action.ToAct == seat {
        t.state.Action.ToAct = t.nextOccupiedFrom(t.state.Action.ToAct)
    }
    // If table becomes empty, reset action.
    if t.state.Seated == 0 {
        t.state.Action = ActionPosition{Dealer: NoSeat, ToAct: NoSeat, HandActive: false}
    }
    return nil
}

func (t *tableImpl) FindPlayer(playerID PlayerID) (SeatIndex, bool) {
    for i := range t.state.Seats {
        s := t.state.Seats[i]
        if s.Status == SeatOccupied && s.Player != nil && s.Player.ID == playerID {
            return SeatIndex(i), true
        }
    }
    return NoSeat, false
}

func (t *tableImpl) StartHand() error {
    if t.state.Seated == 0 {
        return ErrNoPlayers
    }
    t.state.Action.HandActive = true
    // Typically action starts left of dealer preflop; make it generic: next occupied.
    t.state.Action.ToAct = t.nextOccupiedFrom(t.state.Action.Dealer)
    if t.state.Action.ToAct == NoSeat {
        return ErrNoPlayers
    }
    return nil
}

func (t *tableImpl) EndHand() {
    t.state.Action.HandActive = false
    t.state.Action.ToAct = NoSeat
}

func (t *tableImpl) Dealer() SeatIndex { return t.state.Action.Dealer }
func (t *tableImpl) ToAct() SeatIndex { return t.state.Action.ToAct }

func (t *tableImpl) NextToAct() (SeatIndex, error) {
    if !t.state.Action.HandActive {
        return NoSeat, ErrHandNotActive
    }
    next := t.nextOccupiedFrom(t.state.Action.ToAct)
    if next == NoSeat {
        return NoSeat, ErrNoPlayers
    }
    return next, nil
}

func (t *tableImpl) AdvanceAction() (SeatIndex, error) {
    next, err := t.NextToAct()
    if err != nil {
        return NoSeat, err
    }
    t.state.Action.ToAct = next
    return next, nil
}

func (t *tableImpl) AdvanceDealer() error {
    if t.state.Seated == 0 {
        return ErrNoPlayers
    }
    t.state.Action.Dealer = t.nextOccupiedFrom(t.state.Action.Dealer)
    return nil
}

// Helpers

func (t *tableImpl) validSeat(seat SeatIndex) bool {
    return seat >= 0 && int(seat) < t.cfg.MaxSeats
}

// nextOccupiedFrom returns the next occupied seat moving clockwise starting
// after the provided index. Returns NoSeat if none found.
func (t *tableImpl) nextOccupiedFrom(from SeatIndex) SeatIndex {
    if t.state.Seated == 0 {
        return NoSeat
    }
    start := int(from)
    for i := 1; i <= t.cfg.MaxSeats; i++ {
        idx := (start + i) % t.cfg.MaxSeats
        if t.state.Seats[idx].Status == SeatOccupied {
            return SeatIndex(idx)
        }
    }
    return NoSeat
}

func (t *tableImpl) reassignDealer() {
    // Move dealer to next occupied seat or NoSeat if none.
    t.state.Action.Dealer = t.nextOccupiedFrom(t.state.Action.Dealer)
}

