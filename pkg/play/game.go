// Package play runs a cash table with one human seat and heuristic bots.
package play

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/notnil/joker/pkg/bot"
	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/holdem"
)

const (
	defaultSeats = 9
	defaultBuyIn = 10000
)

// Game is a 9-max cash session: one human and eight bots.
type Game struct {
	table *holdem.Table
	hand  *holdem.Hand
	bots  map[int]bot.Player
	hero  int
	buyIn int
	log   []string
}

// NewCash seats a human at hero and fills the rest with mixed-style bots.
func NewCash(hero int, rng *rand.Rand) (*Game, error) {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	if hero < 0 || hero >= defaultSeats {
		hero = 0
	}
	tbl, err := holdem.NewTable(holdem.Config{
		Seats:  defaultSeats,
		Stakes: holdem.Stakes{SmallBlind: 50, BigBlind: 100},
	}, hand.NewDealer(rng))
	if err != nil {
		return nil, err
	}
	styles := []bot.Style{bot.Nit, bot.CallingStation, bot.Lag}
	bots := map[int]bot.Player{}
	names := []string{"Nit", "Station", "LAG"}
	botN := 0
	for seat := range defaultSeats {
		if seat == hero {
			if err := tbl.Sit(seat, "You", defaultBuyIn); err != nil {
				return nil, err
			}
			continue
		}
		st := styles[botN%len(styles)]
		name := fmt.Sprintf("%s-%d", names[botN%len(names)], seat)
		b := bot.New(name, st, rand.New(rand.NewSource(rng.Int63())))
		bots[seat] = b
		if err := tbl.Sit(seat, name, defaultBuyIn); err != nil {
			return nil, err
		}
		botN++
	}
	g := &Game{table: tbl, bots: bots, hero: hero, buyIn: defaultBuyIn}
	if err := g.startHand(); err != nil {
		return nil, err
	}
	g.drainBots()
	return g, nil
}

// View is the human's current snapshot.
func (g *Game) View() holdem.View {
	if g.hand == nil {
		return g.table.View(g.hero)
	}
	return g.hand.View(g.hero)
}

// Hero is the human seat.
func (g *Game) Hero() int { return g.hero }

// Log is a short action history for the UI.
func (g *Game) Log() []string { return append([]string{}, g.log...) }

// Act applies a human action. Illegal actions leave state unchanged.
func (g *Game) Act(a holdem.Action) error {
	if g.hand == nil || g.hand.Done() {
		return holdem.ErrInvalidAction
	}
	if g.hand.ToAct() != g.hero {
		return holdem.ErrInvalidAction
	}
	if err := g.hand.Act(a); err != nil {
		return err
	}
	g.note(g.hero, a)
	return nil
}

// Step performs one bot action, or settles a finished hand and deals the next.
func (g *Game) Step() error {
	if g.hand != nil && g.hand.Done() {
		if err := g.finishHand(); err != nil {
			return err
		}
		if err := g.startHand(); err != nil {
			return err
		}
		return nil
	}
	if g.hand == nil {
		return g.startHand()
	}
	seat := g.hand.ToAct()
	if seat < 0 || seat == g.hero {
		return nil
	}
	return g.botAct(seat)
}

func (g *Game) drainBots() {
	for g.hand != nil && !g.hand.Done() {
		seat := g.hand.ToAct()
		if seat < 0 || seat == g.hero {
			return
		}
		if err := g.botAct(seat); err != nil {
			return
		}
	}
}

func (g *Game) botAct(seat int) error {
	p := g.bots[seat]
	if p == nil {
		return holdem.ErrInvalidAction
	}
	v := g.hand.View(seat)
	a := p.Act(v)
	if err := g.hand.Act(a); err != nil {
		for _, typ := range g.hand.LegalActions() {
			var fb holdem.Action
			switch typ {
			case holdem.Bet:
				fb = holdem.BetAction(g.hand.MinBet())
			case holdem.Raise:
				fb = holdem.RaiseTo(g.hand.MinRaiseTo())
			default:
				var err error
				fb, err = holdem.ParseAction(typ.String(), 0)
				if err != nil {
					continue
				}
			}
			if err := g.hand.Act(fb); err == nil {
				g.note(seat, fb)
				return nil
			}
		}
		return err
	}
	g.note(seat, a)
	return nil
}

func (g *Game) startHand() error {
	h, err := g.table.StartHand()
	if err != nil {
		return err
	}
	g.hand = h
	g.log = append(g.log, "New hand")
	if len(g.log) > 40 {
		g.log = g.log[len(g.log)-40:]
	}
	return nil
}

func (g *Game) finishHand() error {
	if g.hand == nil || !g.hand.Done() {
		return holdem.ErrNoHand
	}
	if err := g.table.ApplyResults(g.hand); err != nil {
		return err
	}
	g.hand = nil
	for seat := range defaultSeats {
		p := g.table.Player(seat)
		if p != nil && p.Chips == 0 {
			_ = g.table.SetChips(seat, g.buyIn)
			g.log = append(g.log, p.ID+" rebuys")
		}
	}
	return nil
}

func (g *Game) note(seat int, a holdem.Action) {
	name := "You"
	if seat != g.hero {
		if b := g.bots[seat]; b != nil {
			name = b.Name()
		}
	}
	msg := name + " " + a.Type.String()
	if a.Chips > 0 {
		msg += fmt.Sprintf(" %d", a.Chips)
	}
	g.log = append(g.log, msg)
	if len(g.log) > 40 {
		g.log = g.log[len(g.log)-40:]
	}
}
