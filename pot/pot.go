package pot

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/loganjspears/joker/hand"
)

// Results is a scam
type Results map[int][]*Result

// Share is the rights a winner has to the pot.
type Share string

const (
	// WonHigh indicates that the high hand was won.
	WonHigh Share = "WonHigh"

	// WonLow indicates that the low hand was won.
	WonLow Share = "WonLow"

	// SplitHigh indicates that the high hand was split.
	SplitHigh Share = "SplitHigh"

	// SplitLow indicates that the low hand was split.
	SplitLow Share = "SplitLow"
)

// A Result is a player's winning result from a showdown.
type Result struct {
	Hand  *hand.Hand `json:"hand"`
	Chips int        `json:"chips"`
	Share Share      `json:"share"`
}

// String returns a string useful for debugging.
func (p *Result) String() string {
	const format = "%s for %d chips with %s"
	return fmt.Sprintf(format, p.Share, p.Chips, p.Hand)
}

// A Pot is the collection of contributions made by players during
// a hand. After the showdown, the pot's chips are divided among the
// winners.
type Pot struct {
	contributions map[int]int
}

// New returns a pot with zero contributions for all seats.
func New(numOfSeats int) *Pot {
	contributions := map[int]int{}
	for i := 0; i < numOfSeats; i++ {
		contributions[i] = 0
	}
	return &Pot{contributions: contributions}
}

// Chips returns the number of chips in the pot.
func (p *Pot) Chips() int {
	chips := 0
	for _, c := range p.contributions {
		chips += c
	}
	return chips
}

// Outstanding returns the amount required for a seat to call the
// largest current bet or raise.
func (p *Pot) Outstanding(seat int) int {
	most := 0
	for _, chips := range p.contributions {
		if chips > most {
			most = chips
		}
	}
	return most - p.contributions[seat]
}

// Contribute contributes the chip amount from the seat given
func (p *Pot) Contribute(seat, chips int) {
	if chips < 0 {
		panic("table: pot contribute negative bet amount")
	}
	p.contributions[seat] += chips
}

// Take creates results with the seat taking the entire pot
func (p *Pot) Take(seat int) Results {
	results := map[int][]*Result{
		seat: []*Result{
			{Hand: nil, Chips: p.Chips(), Share: WonHigh},
		},
	}
	return results
}

// Payout takes the high and low hands to produce pot results.
// Sorting determines how a non-split pot winning hands are sorted.
func (p *Pot) Payout(highHands, lowHands Hands, sorting hand.Sorting, button int) Results {
	sidePots := p.sidePots()
	if len(sidePots) > 1 {
		results := map[int][]*Result{}
		for _, sp := range sidePots {
			r := sp.Payout(highHands, lowHands, sorting, button)
			results = combineResults(results, r)
		}
		return results
	}

	sideHighHands := highHands.handsForSeats(p.seats())
	sideLowHands := lowHands.handsForSeats(p.seats())

	split := len(sideLowHands) > 0
	if !split {
		winners := sideHighHands.WinningHands(sorting)
		switch sorting {
		case hand.SortingHigh:
			return p.resultsFromWinners(winners, p.Chips(), button, highPotShare)
		case hand.SortingLow:
			return p.resultsFromWinners(winners, p.Chips(), button, lowPotShare)
		}
	}

	highWinners := sideHighHands.WinningHands(hand.SortingHigh)
	lowWinners := sideLowHands.WinningHands(hand.SortingLow)

	if len(lowWinners) == 0 {
		return p.resultsFromWinners(highWinners, p.Chips(), button, highPotShare)
	}

	highResults := map[int][]*Result{}
	lowResults := map[int][]*Result{}

	highAmount := p.Chips() / 2
	if highAmount%2 == 1 {
		highAmount++
	}

	highResults = p.resultsFromWinners(highWinners, highAmount, button, highPotShare)
	lowResults = p.resultsFromWinners(lowWinners, p.Chips()/2, button, lowPotShare)
	return combineResults(highResults, lowResults)
}

type potJSON struct {
	Contributions map[string]int
	Chips         int
}

// MarshalJSON conforms to the json.Marshaler interface
func (p *Pot) MarshalJSON() ([]byte, error) {
	m := map[string]int{}
	for seat, chips := range p.contributions {
		seatStr := strconv.FormatInt(int64(seat), 10)
		m[seatStr] = chips
	}

	j := &potJSON{
		Contributions: m,
		Chips:         p.Chips(),
	}
	return json.Marshal(j)
}

// UnmarshalJSON conforms to the json.Marshaler interface
func (p *Pot) UnmarshalJSON(b []byte) error {
	j := &potJSON{}
	if err := json.Unmarshal(b, j); err != nil {
		return err
	}

	m := map[int]int{}
	for seatStr, chips := range j.Contributions {
		seat, err := strconv.ParseInt(seatStr, 10, 64)
		if err != nil {
			return err
		}
		m[int(seat)] = chips
	}

	p.contributions = m
	return nil
}

// resultsFromWinners forms results for winners of the pot
func (p *Pot) resultsFromWinners(winners Hands, chips, button int, f func(n int) Share) map[int][]*Result {
	results := map[int][]*Result{}
	winningSeats := []int{}
	for seat, hand := range winners {
		winningSeats = append(winningSeats, int(seat))
		results[seat] = []*Result{&Result{
			Hand:  hand,
			Chips: chips / len(winners),
			Share: f(len(winners)),
		}}
	}
	sort.IntSlice(winningSeats).Sort()

	remainder := chips % len(winners)
	for i := 0; i < remainder; i++ {
		seatToCheck := (button + i) % 10
		for _, seat := range winningSeats {
			if seat == seatToCheck {
				results[seat][0].Chips++
				break
			}
		}
	}
	return results
}

// sidePots forms an array of side pots including the main pot
func (p *Pot) sidePots() []*Pot {
	// get site pot contribution amounts
	amounts := p.sidePotAmounts()
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

// sidePotAmounts finds the contribution divisions for side pots
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

func highPotShare(n int) Share {
	if n == 1 {
		return WonHigh
	}
	return SplitHigh
}

func lowPotShare(n int) Share {
	if n == 1 {
		return WonLow
	}
	return SplitLow
}

func combineResults(results ...map[int][]*Result) map[int][]*Result {
	combined := map[int][]*Result{}
	for _, m := range results {
		for k, v := range m {
			s := append(combined[k], v...)
			combined[k] = s
		}
	}
	return combined
}
