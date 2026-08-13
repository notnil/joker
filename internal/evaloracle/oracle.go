// Package evaloracle is an independent 5- and 7-card high-hand evaluator used
// to generate testdata corpora. It does not share code with pkg/hand.
package evaloracle

// Ranking values match pkg/hand.Ranking (iota+1).
const (
	HighCard = 1 + iota
	Pair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

// Card is 0–51: rank = c%13 (0=Two … 12=Ace), suit = c/13.
type Card int

func Rank(c Card) int { return int(c) % 13 }
func Suit(c Card) int { return int(c) / 13 }

// Eval5 ranks five cards. Higher score is stronger. Ranking matches pkg/hand.
func Eval5(cards [5]Card) (ranking int, score uint64) {
	var rankN [13]int
	var suitN [4]int
	for _, c := range cards {
		rankN[Rank(c)]++
		suitN[Suit(c)]++
	}
	flush := false
	for _, n := range suitN {
		if n == 5 {
			flush = true
			break
		}
	}

	straight, straightHigh := straightHigh(rankN)
	var quads, trips, pairs []int // ranks, high-to-low
	var kickers []int
	for r := 12; r >= 0; r-- {
		switch rankN[r] {
		case 4:
			quads = append(quads, r)
		case 3:
			trips = append(trips, r)
		case 2:
			pairs = append(pairs, r)
		case 1:
			kickers = append(kickers, r)
		}
	}

	switch {
	case flush && straight && straightHigh == 12:
		return RoyalFlush, pack(RoyalFlush, []int{12})
	case flush && straight:
		return StraightFlush, pack(StraightFlush, []int{straightHigh})
	case len(quads) == 1:
		return FourOfAKind, pack(FourOfAKind, []int{quads[0], kickers[0]})
	case len(trips) == 1 && len(pairs) >= 1:
		return FullHouse, pack(FullHouse, []int{trips[0], pairs[0]})
	case flush:
		return Flush, pack(Flush, kickers)
	case straight:
		return Straight, pack(Straight, []int{straightHigh})
	case len(trips) == 1:
		return ThreeOfAKind, pack(ThreeOfAKind, append([]int{trips[0]}, kickers...))
	case len(pairs) >= 2:
		k := 0
		if len(kickers) > 0 {
			k = kickers[0]
		}
		return TwoPair, pack(TwoPair, []int{pairs[0], pairs[1], k})
	case len(pairs) == 1:
		return Pair, pack(Pair, append([]int{pairs[0]}, kickers...))
	default:
		return HighCard, pack(HighCard, kickers)
	}
}

// Eval7 returns the best 5-card ranking/score from seven cards.
func Eval7(cards [7]Card) (ranking int, score uint64) {
	bestR, bestS := 0, uint64(0)
	choose5(7, func(c [5]int) {
		var five [5]Card
		for i, j := range c {
			five[i] = cards[j]
		}
		r, s := Eval5(five)
		if s > bestS {
			bestR, bestS = r, s
		}
	})
	return bestR, bestS
}

func straightHigh(rankN [13]int) (bool, int) {
	// Ace-high through five-high. Wheel uses high=3 (Five).
	// Presence: rank with at least one card.
	has := func(r int) bool { return rankN[r] > 0 }
	for high := 12; high >= 4; high-- {
		ok := true
		for r := high; r > high-5; r-- {
			if !has(r) {
				ok = false
				break
			}
		}
		if ok {
			return true, high
		}
	}
	if has(12) && has(0) && has(1) && has(2) && has(3) {
		return true, 3 // five-high
	}
	return false, 0
}

func pack(ranking int, kickers []int) uint64 {
	s := uint64(ranking) << 40
	for i, k := range kickers {
		if i >= 5 {
			break
		}
		s |= uint64(k&0xF) << uint(32-4*i)
	}
	return s
}

func choose5(n int, fn func([5]int)) {
	var c [5]int
	var rec func(start, picked int)
	rec = func(start, picked int) {
		if picked == 5 {
			fn(c)
			return
		}
		need := 5 - picked
		for i := start; i <= n-need; i++ {
			c[picked] = i
			rec(i+1, picked+1)
		}
	}
	rec(0, 0)
}
