package table

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/SyntropyDev/joker/hand"
)

// PotShare is the rights a winner has to the pot.
type PotShare string

const (
	// WonHigh indicates that the high hand was won.
	WonHigh PotShare = "Won High"

	// WonLow indicates that the low hand was won.
	WonLow PotShare = "Won Low"

	// SplitHigh indicates that the high hand was split.
	SplitHigh PotShare = "Split High"

	// SplitLow indicates that the low hand was split.
	SplitLow PotShare = "Split Low"
)

// A PotResult is a player's winning result from a showdown.
type PotResult struct {
	Hand  *hand.Hand `json:"hand"`
	Chips int        `json:"chips"`
	Share PotShare   `json:"share"`
}

// String returns a string useful for debugging.
func (p *PotResult) String() string {
	const format = "%s for %d chips with %s"
	return fmt.Sprintf(format, p.Share, p.Chips, p.Hand)
}

// A Pot is the collection of contributions made by players during
// a hand. After the showdown, the pot's chips are divided among the
// winners.
type Pot struct {
	contributions map[int]int
}

// Chips returns the number of chips in the pot.
func (p *Pot) Chips() int {
	chips := 0
	for _, c := range p.contributions {
		chips += c
	}
	return chips
}

type potJSON struct {
	Contributions map[string]int `json:"contributions"`
	Chips         int            `json:"chips"`
}

// MarshalJSON implements the json.Marshaler interface.
func (p *Pot) MarshalJSON() ([]byte, error) {
	contributions := map[string]int{}
	for seat, amount := range p.contributions {
		contributions[strconv.FormatInt(int64(seat), 10)] = amount
	}
	pJSON := &potJSON{
		Chips:         p.Chips(),
		Contributions: contributions,
	}
	return json.Marshal(pJSON)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *Pot) UnmarshalJSON(b []byte) error {
	pJSON := &potJSON{}
	if err := json.Unmarshal(b, pJSON); err != nil {
		return err
	}

	p.contributions = map[int]int{}
	for seat, amount := range pJSON.Contributions {
		i, err := strconv.ParseInt(seat, 10, 64)
		if err != nil {
			return err
		}
		p.contributions[int(i)] = amount
	}

	return nil
}

// Outstanding returns the amount required for a seat to call the
// largest current bet or raise.
func (p *Pot) outstanding(seat int) int {
	most := 0
	for _, chips := range p.contributions {
		if chips > most {
			most = chips
		}
	}
	return most - p.contributions[seat]
}

func newPot(numOfSeats int) *Pot {
	contributions := map[int]int{}
	for i := 0; i < numOfSeats; i++ {
		contributions[i] = 0
	}
	return &Pot{contributions: contributions}
}

func (p *Pot) contribute(seat, chips int) {
	if chips < 0 {
		panic("table: pot contribute negative bet amount")
	}
	p.contributions[seat] += chips
}

func (p *Pot) take(seat int) map[int][]*PotResult {
	results := map[int][]*PotResult{
		seat: []*PotResult{
			{Hand: nil, Chips: p.Chips(), Share: WonHigh},
		},
	}
	return results
}

func (p *Pot) payout(highHands, lowHands tableHands, sorting hand.Sorting, split bool, button int) map[int][]*PotResult {
	// func (p *Pot) payout(highHands, lowHands tableHands, winType winType, button int) map[int][]*PotResult {
	results := map[int][]*PotResult{}
	for _, sidePot := range p.sidePots() {
		highResults := map[int]*PotResult{}
		lowResults := map[int]*PotResult{}
		seats := sidePot.seats()
		sideHighHands := highHands.HandsForSeats(seats)
		sideLowHands := lowHands.HandsForSeats(seats)

		if split {
			highWinners := sideHighHands.WinningHands(winHigh)
			lowWinners := sideLowHands.WinningHands(winLow)
			if len(lowWinners) > 0 {
				highAmount := sidePot.Chips() / 2
				if highAmount%2 == 1 {
					highAmount++
				}
				highResults = resultsFromWinners(highWinners, highAmount, highPotShare)
				lowResults = resultsFromWinners(lowWinners, sidePot.Chips()/2, lowPotShare)
			} else {
				highResults = resultsFromWinners(highWinners, sidePot.Chips(), highPotShare)
			}
		} else {
			switch sorting {
			case hand.SortingHigh:
				winners := sideHighHands.WinningHands(winHigh)
				highResults = resultsFromWinners(winners, sidePot.Chips(), highPotShare)
			case hand.SortingLow:
				winners := sideHighHands.WinningHands(winLow)
				lowResults = resultsFromWinners(winners, sidePot.Chips(), lowPotShare)
			}
		}

		for seat, result := range highResults {
			results[seat] = append(results[seat], result)
		}
		for seat, result := range lowResults {
			results[seat] = append(results[seat], result)
		}
	}
	return results
}

func resultsFromWinners(winners tableHands, chips int, f func(n int) PotShare) map[int]*PotResult {
	results := map[int]*PotResult{}
	winningSeats := []int{}
	for seat, hand := range winners {
		winningSeats = append(winningSeats, int(seat))
		results[seat] = &PotResult{
			Hand:  hand,
			Chips: chips / len(winners),
			Share: f(len(winners)),
		}
	}

	remainder := chips % len(winners)
	sort.IntSlice(winningSeats).Sort()
	for i, seat := range winningSeats {
		if i > remainder {
			results[seat].Chips++
		}
	}
	return results
}

func (p *Pot) sidePots() []*Pot {
	// get site pot contribution amounts
	amounts := p.sidePotAmounts()
	// create side pots
	pots := []*Pot{}
	for i, a := range amounts {
		side := &Pot{
			contributions: map[int]int{},
		}

		last := 0
		if i != 0 {
			last = amounts[i-1]
		}

		for seat, chips := range p.contributions {
			if chips > last && chips >= a {
				side.contributions[seat] = a - last
			} else if chips > last && chips < a {
				side.contributions[seat] = chips - last
			}
		}

		pots = append(pots, side)
	}

	return pots
}

func (p *Pot) sidePotAmounts() []int {
	amounts := []int{}
	for seat, chips := range p.contributions {
		if chips == 0 {
			delete(p.contributions, seat)
		} else {
			found := false
			for _, a := range amounts {
				found = found || a == chips
			}
			if !found {
				amounts = append(amounts, chips)
			}
		}
	}
	sort.IntSlice(amounts).Sort()
	return amounts
}

func (p *Pot) seats() []int {
	seats := []int{}
	for seat := range p.contributions {
		seats = append(seats, seat)
	}
	return seats
}

func highPotShare(n int) PotShare {
	if n == 1 {
		return WonHigh
	}
	return SplitHigh
}

func lowPotShare(n int) PotShare {
	if n == 1 {
		return WonLow
	}
	return SplitLow
}
