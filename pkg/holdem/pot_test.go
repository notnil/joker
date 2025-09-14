package holdem_test

import (
	"testing"

	"github.com/notnil/joker/pkg/holdem"
)

func TestPot(t *testing.T) {
	pot := holdem.NewPot(map[int]int{
		0: 5, 1: 10, 2: 10, 3: 7,
	})
	pot.Remove(3)
	splits := pot.Split()
	if len(splits) != 2 {
		t.Fatalf("expected %d side pots but got %d", 2, len(splits))
	}
	total1 := splits[0].Total()
	if total1 != 20 {
		t.Fatalf("expected %d chips but got %d", 20, total1)
	}
	total2 := splits[1].Total()
	if total2 != 12 {
		t.Fatalf("expected %d chips but got %d", 12, total2)
	}
}

