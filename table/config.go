package table

// An Action is an action a player can take in a hand.
type Action string

const (
	// Fold discards one's hand and forfeits interest in
	// the current pot.
	Fold Action = "Fold"

	// Check is the forfeit to bet when not faced with a bet or
	// raise.
	Check Action = "Check"

	// Call is a match of a bet or raise.
	Call Action = "Call"

	// Bet is a wager that others must match to remain a contender
	// in the current pot.
	Bet Action = "Bet"

	// Raise is an increase to the original bet that others must
	// match to remain a contender in the current pot.
	Raise Action = "Raise"
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

// Stakes are the forced bet amounts for the table.
type Stakes struct {

	// SmallBet is the smaller forced bet amount.
	SmallBet int `json:"smallBet"`

	// BigBet is the bigger forced bet amount.
	BigBet int `json:"bigBet"`

	// Ante is the amount requried from each player to start the hand.
	Ante int `json:"ante"`
}

// Limit is the bet and raise limits of a poker game
type Limit string

const (
	// NoLimit has no limit and players may go "all in"
	NoLimit Limit = "NL"

	// PotLimit has the current value of the pot as the limit
	PotLimit Limit = "PL"

	// FixedLimit restricted the size of bets and raises to predefined
	// values based on the game and round.
	FixedLimit Limit = "FL"
)

// Config are the configurations for creating a table.
type Config struct {

	// Game is the game of the table.
	Game Game `json:"game"`

	// Limit is the limit of the table
	Limit Limit `json:"limit"`

	// Stakes is the stakes for the table.
	Stakes Stakes `json:"stakes"`

	// NumOfSeats is the number of seats available for the table.
	NumOfSeats int `json:"numOfSeats"`
}
