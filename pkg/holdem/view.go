package holdem

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/notnil/joker/pkg/hand"
)

// View is a JSON-friendly snapshot of a hand from one seat's perspective.
// Hole cards are included only for Hero, except at showdown when contesting
// hands are revealed.
type View struct {
	Hero       int         `json:"hero"`
	Seats      int         `json:"seats"`
	Street     string      `json:"street"`
	Board      []hand.Card `json:"board"`
	Pot        int         `json:"pot"`
	ToAct      int         `json:"toAct"`
	Button     int         `json:"button"`
	SB         int         `json:"sb"`
	BB         int         `json:"bb"`
	Call       int         `json:"call"`
	MinBet     int         `json:"minBet"`
	MinRaiseTo int         `json:"minRaiseTo"`
	Legal      []string    `json:"legal"`
	Done       bool        `json:"done"`
	Players    []SeatView  `json:"players"`
}

// SeatView is one seat in a View.
type SeatView struct {
	Seat      int         `json:"seat"`
	ID        string      `json:"id"`
	Chips     int         `json:"chips"`
	StreetPut int         `json:"streetPut"`
	Folded    bool        `json:"folded"`
	AllIn     bool        `json:"allIn"`
	Hole      []hand.Card `json:"hole,omitempty"`
	Hand      string      `json:"hand,omitempty"`
	Award     int         `json:"award,omitempty"`
}

// View returns a snapshot from hero's point of view.
func (h *Hand) View(hero int) View {
	legal := []string{}
	call, minBet, minRaise := 0, 0, 0
	if !h.done && h.toAct == hero {
		for _, a := range h.LegalActions() {
			legal = append(legal, strings.ToLower(a.String()))
		}
		call = h.CallAmount()
		minBet = h.MinBet()
		minRaise = h.MinRaiseTo()
	} else if !h.done {
		minBet = h.MinBet()
		minRaise = h.MinRaiseTo()
		call = h.currentBet
	}

	awards := h.Awards()
	players := make([]SeatView, 0, h.table.config.Seats)
	for seat := range h.table.config.Seats {
		p := h.players[seat]
		if p == nil {
			continue
		}
		sv := SeatView{
			Seat:      seat,
			ID:        p.id,
			Chips:     p.chips,
			StreetPut: p.streetPut,
			Folded:    p.folded,
			AllIn:     p.allIn,
			Award:     awards[seat],
		}
		showHole := seat == hero || (h.done && !p.folded)
		if showHole {
			sv.Hole = slices.Clone(p.hole)
		}
		if h.done && !p.folded {
			for _, r := range h.results {
				if r.Seat == seat && r.Hand != nil {
					sv.Hand = r.Hand.Description()
				}
			}
		}
		players = append(players, sv)
	}

	street := ""
	if h.done {
		street = "Showdown"
		if len(h.contesting()) <= 1 {
			street = "Fold"
		}
	} else {
		street = h.street.String()
	}

	return View{
		Hero:       hero,
		Seats:      h.table.config.Seats,
		Street:     street,
		Board:      h.Board(),
		Pot:        h.pot.Total(),
		ToAct:      h.toAct,
		Button:     h.button,
		SB:         h.sb,
		BB:         h.bb,
		Call:       call,
		MinBet:     minBet,
		MinRaiseTo: minRaise,
		Legal:      legal,
		Done:       h.done,
		Players:    players,
	}
}

// TableView is seating between hands (no hole cards).
func (t *Table) View(hero int) View {
	players := make([]SeatView, 0, t.config.Seats)
	for seat, p := range t.seats {
		if p == nil {
			continue
		}
		players = append(players, SeatView{
			Seat:  seat,
			ID:    p.ID,
			Chips: p.Chips,
		})
	}
	return View{
		Hero:    hero,
		Seats:   t.config.Seats,
		Street:  "",
		ToAct:   -1,
		Button:  t.button,
		Players: players,
	}
}

// SetChips sets a seated player's stack (cash rebuy).
func (t *Table) SetChips(seat, chips int) error {
	if seat < 0 || seat >= t.config.Seats || t.seats[seat] == nil {
		return ErrInvalidSeat
	}
	if chips < 0 {
		return ErrInvalidAmount
	}
	t.seats[seat].Chips = chips
	return nil
}

// PlayerID returns the player id at seat in this hand.
func (h *Hand) PlayerID(seat int) string {
	p := h.players[seat]
	if p == nil {
		return ""
	}
	return p.id
}

// ParseAction maps UI/bot strings to an Action.
func ParseAction(typ string, chips int) (Action, error) {
	switch strings.ToLower(strings.TrimSpace(typ)) {
	case "fold":
		return FoldAction(), nil
	case "check":
		return CheckAction(), nil
	case "call":
		return CallAction(), nil
	case "bet":
		return BetAction(chips), nil
	case "raise":
		return RaiseTo(chips), nil
	case "allin", "all-in", "all_in":
		return AllInAction(), nil
	default:
		return Action{}, ErrInvalidAction
	}
}

func (a ActionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToLower(a.String()))
}
