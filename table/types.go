package table

// SeatIndex represents a position at the table.
// -1 means "no seat" or undefined.
type SeatIndex int

const (
    NoSeat SeatIndex = -1
)

// PlayerID is a unique identifier for a player.
type PlayerID string

// Player represents basic identity info used by the table.
type Player struct {
    ID   PlayerID
    Name string
}

// SeatStatus represents occupancy state for a seat.
type SeatStatus int

const (
    SeatEmpty SeatStatus = iota
    SeatOccupied
)

// Seat models a single position at the table.
type Seat struct {
    Index  SeatIndex
    Status SeatStatus
    Player *Player
}

// ActionPosition indicates who is to act.
type ActionPosition struct {
    Dealer     SeatIndex
    ToAct      SeatIndex
    HandActive bool
}

// Config encapsulates static table configuration.
type Config struct {
    MaxSeats int
}

// State captures dynamic table state.
type State struct {
    Seats   []Seat
    Action  ActionPosition
    Seated  int
}

// Errors returned by table operations.
var (
    ErrInvalidSeat   = newError("invalid seat index")
    ErrSeatOccupied  = newError("seat already occupied")
    ErrSeatEmpty     = newError("seat empty")
    ErrTableFull     = newError("table full")
    ErrPlayerSeated  = newError("player already seated")
    ErrNoPlayers     = newError("no players seated")
    ErrHandNotActive = newError("hand not active")
)

type tableError struct{ msg string }

func (e tableError) Error() string { return e.msg }

func newError(msg string) error { return tableError{msg: msg} }

