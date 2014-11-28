package table

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ErrorType represents an enum value of one of the possible
// error types.
type ErrorType int

const (
	// InvalidBuyIn errors occur when a player attempts to sit at a
	// table with an invalid buyin.
	InvalidBuyIn ErrorType = iota + 1

	// SeatOccupied errors occur when a player attempts to sit at a
	// table in a seat that is already occupied.
	SeatOccupied

	// InvalidSeat errors occur when a player attempts to sit at a
	// table in a seat that is invalid.
	InvalidSeat

	// AlreadySeated errors occur when a player attempts to sit at a
	// table at which the player is already seated.
	AlreadySeated

	// InsufficientPlayers errors occur when the table's Next() method
	// can't start a new hand because of insufficient players
	InsufficientPlayers

	// InvalidBet errors occur when a player attempts to bet an invalid
	// amount.  Bets are invalid if they exceed a player's chips or fall below the
	// stakes minimum bet.  In fixed limit games the bet amount must equal the amount
	// prespecified by the limit and round.  In pot limit games the bet must be less
	// than or equal to the pot.
	InvalidBet

	// InvalidRaise errors occur when a player attempts to raise an invalid
	// amount.  Raises are invalid if the raise or reraise is lower than the previous bet
	// or raised amount unless it puts the player allin.  Raises are also invalid if they
	// exceed a player's chips. In fixed limit games the raise amount must equal the amount
	// prespecified by the limit and round.  In pot limit games the raise must be less
	// than or equal to the pot.
	InvalidRaise

	// InvalidAction errors occur when a player attempts an action that isn't
	// currently allowed.  For example a check action is invalid when faced with a raise.
	InvalidAction
)

func (e ErrorType) String() string {
	switch e {
	case InvalidBuyIn:
		return "Invalid Buy In"
	case SeatOccupied:
		return "Seat Occupied"
	case InvalidSeat:
		return "Invalid Seat"
	case AlreadySeated:
		return "Already Seated"
	case InsufficientPlayers:
		return "Insufficient Players"
	case InvalidBet:
		return "Invalid Bet"
	case InvalidRaise:
		return "Invalid Raise"
	case InvalidAction:
		return "Invalid Action"
	}
	return ""
}

// An Error contains an ErrorType and a error string.  It conforms to the error interface.
type Error struct {
	errType ErrorType
	errStr  string
}

// NewInvalidBuyIn returns an error with the InvalidBuyIn type
func NewInvalidBuyIn(chips int) *Error {
	return &Error{
		errType: InvalidBuyIn,
		errStr:  fmt.Sprintf("table: %d is an invalid buy in amount", chips),
	}
}

// NewSeatOccupied returns an error with the SeatOccupied type
func NewSeatOccupied(seat int) *Error {
	return &Error{
		errType: SeatOccupied,
		errStr:  fmt.Sprintf("table: seat %d is occupied", seat),
	}
}

// NewInvalidSeat returns an error with the InvalidSeat type
func NewInvalidSeat(seat int) *Error {
	return &Error{
		errType: InvalidSeat,
		errStr:  fmt.Sprintf("table: seat %d is invalid", seat),
	}
}

// NewAlreadySeated returns an error with the AlreadySeated type
func NewAlreadySeated(id string) *Error {
	return &Error{
		errType: AlreadySeated,
		errStr:  fmt.Sprintf("table: player with id %s is already seated", id),
	}
}

// NewInsufficientPlayers returns an error with the InsufficientPlayers type
func NewInsufficientPlayers() *Error {
	return &Error{
		errType: InsufficientPlayers,
		errStr:  fmt.Sprintf("table: not enough players to start hand"),
	}
}

// NewInvalidBet returns an error with the InvalidBet type
func NewInvalidBet(bet int, min int, max int) *Error {
	return &Error{
		errType: InvalidBet,
		errStr:  fmt.Sprintf("table: invalid bet of %d, must be between %d and %d", bet, min, max),
	}
}

// NewInvalidRaise returns an error with the InvalidRaise type
func NewInvalidRaise(bet int, min int, max int) *Error {
	return &Error{
		errType: InvalidRaise,
		errStr:  fmt.Sprintf("table: invalid raise of %d, must be between %d and %d", bet, min, max),
	}
}

// NewInvalidAction returns an error with the InvalidAction type
func NewInvalidAction(invalid Action, validActions []Action) *Error {
	actionStrs := []string{}
	for _, a := range validActions {
		actionStrs = append(actionStrs, string(a))
	}

	format := "table: player attempted invalid action %s, valid actions are - "
	return &Error{
		errType: InvalidAction,
		errStr:  fmt.Sprintf(format, invalid, strings.Join(actionStrs, ",")),
	}
}

// Type returns the ErrorType of the Error.
func (e *Error) Type() ErrorType {
	return e.errType
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.errStr
}

type errorJSON struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

// MarshalJSON implements the json.Marshaler interface.
// {"type":"Invalid Buy In","error":"table: 5 is an invalid buy in amount"}
func (e *Error) MarshalJSON() (b []byte, err error) {
	eJSON := &errorJSON{
		Type:  e.errType.String(),
		Error: e.Error(),
	}
	return json.Marshal(eJSON)
}
