package hand_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/jokertest"
)

type holdem7Row struct {
	Cards   []string `json:"cards"`
	Ranking int      `json:"ranking"`
	Score   uint64   `json:"score"`
}

func TestHoldem7Corpus(t *testing.T) {
	path := corpusPath(t, "eval", "holdem7.jsonl")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	n := 0
	var prev *hand.Hand
	var prevScore uint64
	for sc.Scan() {
		var row holdem7Row
		if err := json.Unmarshal(sc.Bytes(), &row); err != nil {
			t.Fatalf("line %d: %v", n+1, err)
		}
		cards := jokertest.Cards(row.Cards...)
		h := hand.New(cards)
		if int(h.Ranking()) != row.Ranking {
			t.Errorf("line %d %v: ranking %s (%d), want %d", n+1, row.Cards, h.Ranking(), h.Ranking(), row.Ranking)
		}
		if prev != nil {
			cmp := h.CompareTo(prev)
			switch {
			case row.Score > prevScore && cmp <= 0:
				t.Errorf("line %d: score increased but CompareTo=%d", n+1, cmp)
			case row.Score < prevScore && cmp >= 0:
				t.Errorf("line %d: score decreased but CompareTo=%d", n+1, cmp)
			case row.Score == prevScore && cmp != 0:
				t.Errorf("line %d: score tied but CompareTo=%d", n+1, cmp)
			}
		}
		prev = h
		prevScore = row.Score
		n++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if n < 1000 {
		t.Fatalf("corpus too small: %d", n)
	}
}

func corpusPath(t *testing.T, parts ...string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(append([]string{root, "testdata"}, parts...)...)
}
