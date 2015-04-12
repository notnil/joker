package table

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Player represents a player at a table.
type Player interface {
	// ID returns the unique identifier of the player.
	ID() string

	// FromID resets the player from an id.  It is required for
	// deserialization.
	FromID(id string) (Player, error)

	// Action returns the action and it's chip amount.  This method
	// will block table's Next() function until input is recieved.
	Action() (a Action, chips int)
}

// RegisterPlayer stores the player implementation for json deserialization.
func RegisterPlayer(p Player) {
	registeredPlayer = p
}

var (
	// mapping to player implemenation
	registeredPlayer Player
)

// PlayerState includes table information about a player
type PlayerState struct {
	player    Player
	holeCards []*HoleCard
	chips     int
	acted     bool
	out       bool
	allin     bool
	canRaise  bool
}

// Acted returns whether or not the player has acted for the current round.
func (state *PlayerState) Acted() bool {
	return state.acted
}

// AllIn returns whether or not the player is all in for the current hand.
func (state *PlayerState) AllIn() bool {
	return state.allin
}

// CanRaise returns whether or not the player can raise in the current round.
func (state *PlayerState) CanRaise() bool {
	return state.canRaise
}

// Chips returns the number of chips the player has in his or her stack.
func (state *PlayerState) Chips() int {
	return state.chips
}

// HoleCards returns the hole cards the player currently has.
func (state *PlayerState) HoleCards() []*HoleCard {
	c := []*HoleCard{}
	return append(c, state.holeCards...)
}

// Out returns whether or not the player is out of the current hand.
func (state *PlayerState) Out() bool {
	return state.out
}

// Player returns the player.
func (state *PlayerState) Player() Player {
	return state.player
}

// String returns a string useful for debugging.
func (state *PlayerState) String() string {
	const format = "{Player: %s, HoleCards: %v, Chips: %d, Acted: %t, Out: %t, AllIn: %t}"
	return fmt.Sprintf(format,
		state.player.ID(), state.holeCards, state.chips, state.acted, state.out, state.allin)
}

type playerStateJSON struct {
	ID        string      `json:"id"`
	HoleCards []*HoleCard `json:"holeCards"`
	Chips     int         `json:"chips"`
	Acted     bool        `json:"acted"`
	Out       bool        `json:"out"`
	Allin     bool        `json:"allin"`
	CanRaise  bool        `json:"canRaise"`
}

// MarshalJSON implements the json.Marshaler interface.
func (state *PlayerState) MarshalJSON() ([]byte, error) {
	tpJSON := &playerStateJSON{
		ID:        state.Player().ID(),
		HoleCards: state.HoleCards(),
		Chips:     state.Chips(),
		Acted:     state.Acted(),
		Out:       state.Out(),
		Allin:     state.AllIn(),
		CanRaise:  state.CanRaise(),
	}
	return json.Marshal(tpJSON)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (state *PlayerState) UnmarshalJSON(b []byte) error {
	tpJSON := &playerStateJSON{}
	if err := json.Unmarshal(b, tpJSON); err != nil {
		return err
	}

	if isNil(registeredPlayer) {
		return errors.New("table: PlayerState json deserialization requires use of the RegisterPlayer function")
	}

	p, err := registeredPlayer.FromID(tpJSON.ID)
	if err != nil {
		return fmt.Errorf("table PlayerState json deserialization failed because of player %s FromID - %s", tpJSON.ID, err)
	}

	state.player = p
	state.holeCards = tpJSON.HoleCards
	state.chips = tpJSON.Chips
	state.acted = tpJSON.Acted
	state.out = tpJSON.Out
	state.allin = tpJSON.Allin
	state.canRaise = tpJSON.CanRaise

	return nil
}
