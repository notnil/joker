package holdem

import "slices"

// ActionType is a legal betting action.
type ActionType int

const (
	Fold ActionType = iota
	Check
	Call
	Bet
	Raise
	AllIn
)

var actionNames = []string{"Fold", "Check", "Call", "Bet", "Raise", "AllIn"}

func (a ActionType) String() string {
	if a < 0 || int(a) >= len(actionNames) {
		return "ActionType(?)"
	}
	return actionNames[a]
}

// Action is a betting action. For Bet, Chips is the bet size. For Raise,
// Chips is the total contribution this street (raise-to).
type Action struct {
	Type  ActionType
	Chips int
}

func FoldAction() Action  { return Action{Type: Fold} }
func CheckAction() Action { return Action{Type: Check} }
func CallAction() Action  { return Action{Type: Call} }
func BetAction(chips int) Action {
	return Action{Type: Bet, Chips: chips}
}
func RaiseTo(chips int) Action {
	return Action{Type: Raise, Chips: chips}
}
func AllInAction() Action { return Action{Type: AllIn} }

func includesType(types []ActionType, t ActionType) bool {
	return slices.Contains(types, t)
}
