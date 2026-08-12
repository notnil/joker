package evaloracle

import "testing"

func TestOracleFiveCardCensus(t *testing.T) {
	counts := map[int]int{}
	n := 0
	for a := 0; a < 52; a++ {
		for b := a + 1; b < 52; b++ {
			for c := b + 1; c < 52; c++ {
				for d := c + 1; d < 52; d++ {
					for e := d + 1; e < 52; e++ {
						r, _ := Eval5([5]Card{Card(a), Card(b), Card(c), Card(d), Card(e)})
						counts[r]++
						n++
					}
				}
			}
		}
	}
	if n != 2598960 {
		t.Fatalf("enumerated %d", n)
	}
	want := map[int]int{
		HighCard:      1302540,
		Pair:          1098240,
		TwoPair:       123552,
		ThreeOfAKind:  54912,
		Straight:      10200,
		Flush:         5108,
		FullHouse:     3744,
		FourOfAKind:   624,
		StraightFlush: 36,
		RoyalFlush:    4,
	}
	for r, w := range want {
		if counts[r] != w {
			t.Errorf("ranking %d: got %d, want %d", r, counts[r], w)
		}
	}
}
