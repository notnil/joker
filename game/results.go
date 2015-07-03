package game

import "github.com/loganjspears/joker/hand"

// Results is a mapping of player's seat to Result.
type Results map[int][]*Result

func (r Results) merge(o Results) Results {
	c := map[int][]*Result{}
	for seat, results := range r {
		c[seat] = append(results, o[seat]...)
	}
	return Results(c)
}

// Share is the rights a winner has to the pot.
type Share int

const (
	// WonHigh indicates that the high hand was won.
	WonHigh Share = iota + 1

	// WonLow indicates that the low hand was won.
	WonLow

	// SplitHigh indicates that the high hand was split.
	SplitHigh

	// SplitLow indicates that the low hand was split.
	SplitLow
)

// A Result is a player's winning result from a showdown.
type Result struct {
	hand  *hand.Hand
	share Share
}

// Hand returns the showdown result's playing hand.
func (r *Result) Hand() *hand.Hand {
	return r.hand
}

// Share returns the result's share of the pot.
func (r *Result) Share() Share {
	return r.share
}

func shareForSorting(sort hand.Sorting, n int) Share {
	switch sort {
	case hand.SortingHigh:
		switch n {
		case 1:
			return WonHigh
		default:
			return SplitHigh
		}
	case hand.SortingLow:
		switch n {
		case 1:
			return WonLow
		default:
			return SplitLow
		}
	}
	return WonHigh
}
