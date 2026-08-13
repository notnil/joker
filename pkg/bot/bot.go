// Package bot implements simple No-Limit Hold'em heuristic players.
package bot

import (
	"math/rand"

	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/holdem"
)

// Style is a betting personality.
type Style int

const (
	Nit Style = iota
	CallingStation
	Lag
)

func (s Style) String() string {
	switch s {
	case Nit:
		return "Nit"
	case CallingStation:
		return "Station"
	case Lag:
		return "LAG"
	default:
		return "Bot"
	}
}

// Player chooses a legal action from a hidden-card view.
type Player interface {
	Name() string
	Act(v holdem.View) holdem.Action
}

type heuristic struct {
	name  string
	style Style
	rng   *rand.Rand
}

// New returns a heuristic bot. rng may be nil (a default source is used).
func New(name string, style Style, rng *rand.Rand) Player {
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}
	return &heuristic{name: name, style: style, rng: rng}
}

func (h *heuristic) Name() string { return h.name }

func (h *heuristic) Act(v holdem.View) holdem.Action {
	legal := map[string]bool{}
	for _, a := range v.Legal {
		legal[a] = true
	}
	pick := func(typ string, chips int) holdem.Action {
		a, err := holdem.ParseAction(typ, chips)
		if err != nil || !legal[typ] {
			return holdem.Action{}
		}
		return a
	}
	fallback := func() holdem.Action {
		for _, typ := range []string{"check", "call", "fold", "allin"} {
			if legal[typ] {
				a, _ := holdem.ParseAction(typ, 0)
				return a
			}
		}
		return holdem.FoldAction()
	}

	if len(v.Legal) == 0 {
		return holdem.FoldAction()
	}

	score := strength(v)
	agg := aggression(h.style)
	callThresh, raiseThresh := thresholds(h.style)

	if legal["check"] && score < callThresh && h.rng.Float64() > agg*0.3 {
		return pick("check", 0)
	}

	if score >= raiseThresh && h.rng.Float64() < agg {
		if legal["raise"] {
			return pick("raise", raiseTo(v, h.rng))
		}
		if legal["bet"] {
			return pick("bet", betSize(v, h.rng))
		}
		if legal["allin"] && score > 0.85 {
			return pick("allin", 0)
		}
	}

	if score >= callThresh {
		if legal["call"] {
			return pick("call", 0)
		}
		if legal["check"] {
			return pick("check", 0)
		}
	}

	if legal["check"] {
		return pick("check", 0)
	}
	if legal["fold"] && score < callThresh {
		return pick("fold", 0)
	}
	return fallback()
}

func aggression(s Style) float64 {
	switch s {
	case Nit:
		return 0.15
	case CallingStation:
		return 0.2
	case Lag:
		return 0.7
	default:
		return 0.4
	}
}

func thresholds(s Style) (call, raise float64) {
	switch s {
	case Nit:
		return 0.62, 0.8
	case CallingStation:
		return 0.22, 0.78
	case Lag:
		return 0.32, 0.5
	default:
		return 0.4, 0.7
	}
}

func strength(v holdem.View) float64 {
	var hole []hand.Card
	for _, p := range v.Players {
		if p.Seat == v.Hero {
			hole = p.Hole
			break
		}
	}
	if len(hole) < 2 {
		return 0
	}
	if len(v.Board) == 0 {
		return preflop(hole)
	}
	cards := append(append([]hand.Card{}, hole...), v.Board...)
	made := hand.New(cards)
	switch made.Ranking() {
	case hand.RoyalFlush, hand.StraightFlush:
		return 1
	case hand.FourOfAKind:
		return 0.98
	case hand.FullHouse:
		return 0.92
	case hand.Flush:
		return 0.85
	case hand.Straight:
		return 0.8
	case hand.ThreeOfAKind:
		return 0.72
	case hand.TwoPair:
		return 0.6
	case hand.Pair:
		return 0.45
	default:
		return 0.2 + preflop(hole)*0.15
	}
}

func preflop(hole []hand.Card) float64 {
	a, b := hole[0], hole[1]
	r1, r2 := a.Rank(), b.Rank()
	if r1 < r2 {
		r1, r2 = r2, r1
	}
	if r1 == r2 {
		return 0.55 + float64(r1)/26.0
	}
	s := float64(r1)/13.0*0.45 + float64(r2)/13.0*0.15
	if a.Suit() == b.Suit() {
		s += 0.12
	}
	gap := int(r1 - r2)
	if gap == 1 {
		s += 0.1
	} else if gap == 2 {
		s += 0.04
	}
	if r1 == hand.Ace {
		s += 0.08
	}
	if s > 0.95 {
		s = 0.95
	}
	return s
}

func heroChips(v holdem.View) (chips, streetPut int) {
	for _, p := range v.Players {
		if p.Seat == v.Hero {
			return p.Chips, p.StreetPut
		}
	}
	return 0, 0
}

func betSize(v holdem.View, rng *rand.Rand) int {
	chips, _ := heroChips(v)
	min := v.MinBet
	if min < 1 {
		min = 1
	}
	if chips <= min {
		return chips
	}
	switch rng.Intn(3) {
	case 0:
		return min
	case 1:
		sz := min * 5 / 2
		if sz > chips {
			return chips
		}
		if sz < min {
			return min
		}
		return sz
	default:
		return chips
	}
}

func raiseTo(v holdem.View, rng *rand.Rand) int {
	chips, street := heroChips(v)
	maxTo := street + chips
	minTo := v.MinRaiseTo
	if minTo < 1 {
		minTo = v.Call + v.MinBet
	}
	if minTo > maxTo {
		return maxTo
	}
	switch rng.Intn(3) {
	case 0:
		return minTo
	case 1:
		sz := minTo + (maxTo-minTo)/2
		if sz < minTo {
			return minTo
		}
		return sz
	default:
		return maxTo
	}
}
