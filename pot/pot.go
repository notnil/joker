package pot

import (
	"encoding/json"
	"sort"

	"github.com/notnil/joker/pot"
)

const (
	noPosToAct = -1
)

type Action int

const (
	Fold Action = iota
	Check
	Call
	Bet
	Raise
)

type Seat struct {
	Pos         int
	Stack       int
	Contributed int
	Acted       bool
	Folded      bool
	AllIn       bool
}

func (s *Seat) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *Seat) copy() *Seat {
	return &Seat{
		Pos:         s.Pos,
		Stack:       s.Stack,
		Contributed: s.Contributed,
		Acted:       s.Acted,
		Folded:      s.Folded,
		AllIn:       s.AllIn,
	}
}

type Pot struct {
	seats    []*Seat
	posToAct int
	cost     int
	err      error
}

func Blinds(button, small, big int) func(p *Pot) {
	// TODO catch invalid button
	return func(p *Pot) {
		p.posToAct = button
		blinds := []int{small, big}
		if len(p.seats) == 2 {
			for _, blind := range blinds {
				p.contribute(p.SeatToAct(), blind, false)
				p.update()
			}
		} else {
			for _, blind := range blinds {
				p.update()
				p.contribute(p.SeatToAct(), blind, false)
			}
		}
	}
}

func Ante(chips int) func(p *Pot) {
	return func(p *Pot) {
		for _, seat := range p.seats {
			p.contribute(seat, chips, false)
		}
	}
}

func New(stacks map[int]int, opts ...func(*Pot)) *Pot {
	// TODO catch only one seat given
	seats := []*Seat{}
	for seat, stack := range stacks {
		seats = append(seats, &Seat{Pos: seat, Stack: stack})
	}
	sort.Slice(seats, func(i, j int) bool {
		return seats[i].Pos < seats[j].Pos
	})
	p := &Pot{
		seats:    seats,
		posToAct: seats[0].Pos,
	}
	for _, f := range opts {
		f(p)
	}
	return p
}

func (p *Pot) Chips() int {
	total := 0
	for _, seat := range p.seats {
		total += seat.Contributed
	}
	return total
}

func (p *Pot) Seats() []*Seat {
	return append([]*Seat{}, p.seats...)
}

func (p *Pot) SeatToAct() *Seat {
	if p.posToAct == noPosToAct {
		return nil
	}
	return p.seats[p.posToAct]
}

func (p *Pot) PossibleActions() []Action {
	seat := p.SeatToAct()
	if seat == nil {
		return []Action{}
	}
	if p.cost == 0 {
		return []Action{Fold, Check, Bet}
	}
	if p.cost >= seat.Stack {
		return []Action{Fold, Call}
	}
	return []Action{Fold, Call, Raise}
}

func (p *Pot) Fold() *Pot {
	p.SeatToAct().Folded = true
	p.update()
	return p
}

func (p *Pot) Check() {
	if includes(p.PossibleActions(), Check) == false {

	}
	p.SeatToAct().Folded = true
	p.update()
}

func (p *Pot) update() {
	p.moveAction()
	p.setCost()
}

func (p *Pot) moveAction() {
	if p.posToAct == noPosToAct {
		return
	}
	for i := 1; i < len(p.seats); i++ {
		a := (p.posToAct + i) % len(p.seats)
		if p.seats[a].Folded == false && p.seats[a].AllIn == false && p.seats[a].Acted == false {
			p.posToAct = a
			return
		}
	}
	p.posToAct = noPosToAct
}

func (p *Pot) setCost() {
	if p.posToAct == noPosToAct {
		p.cost = 0
		return
	}
	highest := 0
	for _, seat := range p.seats {
		if seat.Contributed > highest {
			highest = seat.Contributed
		}
	}
	p.cost = highest - p.SeatToAct().Contributed
}

func (p *Pot) contribute(seat *Seat, chips int, acted bool) {
	stack := seat.Stack
	amount := chips
	if stack <= chips {
		amount = stack
		seat.AllIn = true
	}
	seat.Contributed += amount
	seat.Stack -= amount
	seat.Acted = acted
}

func (p *Pot) copy() *Pot {
	seats := []*Seat{}
	for _, seat := range p.seats {
		seats = append(seats, seat.copy())
	}
	return &Pot{
		seats:    seats,
		posToAct: p.posToAct,
		cost:     cost,
	}
}

func includes(actions []pot.Action, include ...pot.Action) bool {
	for _, a1 := range include {
		found := false
		for _, a2 := range actions {
			found = found || a1 == a2
		}
		if !found {
			return false
		}
	}
	return true
}

// // A Share is the share of the pot a seat is entitled to.
// type Share struct {
// 	Chips int
// 	Type  ShareType
// }

// // String returns a string useful for debugging.
// func (s *Share) String() string {
// 	const format = "%s for %d chips"
// 	return fmt.Sprintf(format, s.Type, s.Chips)
// }

// type Pot struct {
// 	contributions map[int]int
// 	in            map[int]bool
// 	chips         int
// 	err           error
// }

// // New returns a pot.
// func New(seats []int) *Pot {
// 	c := map[int]int{}
// 	in := map[int]bool{}
// 	for _, seat := range seats {
// 		c[seat] = 0
// 		in[seat] = true
// 	}
// 	return &Pot{contributions: c, in: map[int]bool{}, chips: 0}
// }

// // Error returns an error if the pot is in an invalid state
// func (p *Pot) Error() error {
// 	return p.err
// }

// // Chips returns the number of chips in the pot.
// func (p *Pot) Chips() int {
// 	return p.chips
// }

// // Outstanding returns the amount required for a seat to call the
// // largest current bet or raise.
// func (p *Pot) Outstanding(seat int) int {
// 	most := 0
// 	for _, chips := range p.contributions {
// 		if chips > most {
// 			most = chips
// 		}
// 	}
// 	return most - p.contributions[seat]
// }

// // Contribute contributes the chip amount from the seat given
// func (p *Pot) Contribute(seat, chips int, allin bool) *Pot {
// 	p.checkSeats([]int{seat})
// }

// // Withdrawl withdrawls the seat given from the pot.  This places
// // the seat out of contention for shares of the pot.
// func (p *Pot) Withdrawl(seat int) {
// 	p.checkSeats([]int{seat})
// 	p.in[seat] = false
// }

// // Fold withdrawls the seat given from the pot.  This
// func (p *Pot) Contested() bool {
// 	p.checkSeats([]int{seat})
// 	p.in[seat] = false
// }

// // Showdown takes the high and low hands to produce pot results.
// // Highs and lows represent showdowns for high and low portions of
// // the pot.  If the pot isn't split only use highs.  Highs and lows
// // should be order in descreasing order of claim to the pot.
// func (p *Pot) Showdown(highs, lows []int, button int) *Pot {
// 	// defend against invalid claims
// 	p.checkSeats(append(highs, lows...))
// 	// don't continue if the pot is invalid
// 	if p.err != nil {
// 		return p
// 	}
// }

// // sidePots forms an array of side pots including the main pot
// func (p *Pot) sidePots() []*Pot {
// 	// get site pot contribution amounts
// 	amounts := p.sidePotAmounts()
// 	pots := []*Pot{}
// 	for i, a := range amounts {
// 		side := &Pot{
// 			contributions: map[int]int{},
// 		}
// 		last := 0
// 		if i != 0 {
// 			last = amounts[i-1]
// 		}
// 		for seat, chips := range p.contributions {
// 			if chips > last && chips >= a {
// 				side.contributions[seat] = a - last
// 			} else if chips > last && chips < a {
// 				side.contributions[seat] = chips - last
// 			}
// 		}
// 		pots = append(pots, side)
// 	}
// 	return pots
// }

// // sidePotAmounts finds the contribution divisions for side pots
// func (p *Pot) sidePotAmounts() []int {
// 	amounts := []int{}
// 	for seat, chips := range p.contributions {
// 		if chips == 0 {
// 			delete(p.contributions, seat)
// 		} else {
// 			found := false
// 			for _, a := range amounts {
// 				found = found || a == chips
// 			}
// 			if !found {
// 				amounts = append(amounts, chips)
// 			}
// 		}
// 	}
// 	sort.IntSlice(amounts).Sort()
// 	return amounts
// }

// func (p *Pot) seats() []int {
// 	seats := []int{}
// 	for seat := range p.contributions {
// 		seats = append(seats, seat)
// 	}
// 	return seats
// }

// func (p *Pot) checkSeats(seats []int) {
// 	for _, seat := range seats {
// 		if p.in[seat] == false {
// 			p.err = errors.New("pot: seat attempted showdown but not in pot")
// 			return
// 		}
// 	}
// }

// func (p *Pot) seatsRemaining() int {
// 	count := 0
// 	for _, in := range p.in {
// 		if in {
// 			count++
// 		}
// 	}
// 	return count
// }

// func combineResults(results ...map[int][]*Result) map[int][]*Result {
// 	combined := map[int][]*Result{}
// 	for _, m := range results {
// 		for k, v := range m {
// 			s := append(combined[k], v...)
// 			combined[k] = s
// 		}
// 	}
// 	return combined
// }
