package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/notnil/joker/internal/evaloracle"
)

var ranks = []byte("23456789TJQKA")
var suits = []byte("shdc")

type row struct {
	Cards   []string `json:"cards"`
	Ranking int      `json:"ranking"`
	Score   uint64   `json:"score"`
}

func main() {
	outDir := "testdata/eval"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	path := filepath.Join(outDir, "holdem7.jsonl")
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	seen := map[[7]evaloracle.Card]bool{}
	write := func(cs [7]evaloracle.Card) {
		if seen[cs] {
			return
		}
		seen[cs] = true
		r, s := evaloracle.Eval7(cs)
		row := row{Cards: format7(cs), Ranking: r, Score: s}
		b, err := json.Marshal(row)
		if err != nil {
			panic(err)
		}
		if _, err := w.Write(append(b, '\n')); err != nil {
			panic(err)
		}
	}

	// Famous / edge 7-card patterns (remaining cards are fillers).
	write(must7("As", "Ks", "Qs", "Js", "Ts", "2h", "3d")) // royal
	write(must7("Ah", "5h", "4h", "3h", "2h", "Kd", "Qc")) // steel wheel
	write(must7("9s", "9h", "9d", "9c", "As", "Kh", "Qc")) // quads
	write(must7("As", "Ah", "Ad", "Ks", "Kh", "2c", "3d")) // full house
	write(must7("As", "2h", "3d", "4c", "5s", "9h", "9c")) // wheel vs pair on board
	write(must7("Kc", "Qc", "Jc", "Tc", "9c", "Ah", "2d")) // king-high SF
	write(must7("Ah", "Ad", "Kc", "Kd", "2s", "3h", "4c")) // two pair
	write(must7("7s", "8s", "9s", "Ts", "2s", "3d", "4c")) // flush not straight
	write(must7("As", "Ks", "Qs", "Js", "9s", "Th", "2d")) // broadway straight, flush miss

	rng := rand.New(rand.NewSource(20260812))
	for len(seen) < 4000 {
		write(deal7(rng))
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d hands to %s\n", len(seen), path)
}

func deal7(rng *rand.Rand) [7]evaloracle.Card {
	perm := rng.Perm(52)
	var cs [7]evaloracle.Card
	for i := 0; i < 7; i++ {
		cs[i] = evaloracle.Card(perm[i])
	}
	return cs
}

func format7(cs [7]evaloracle.Card) []string {
	out := make([]string, 7)
	for i, c := range cs {
		out[i] = fmt.Sprintf("%c%c", ranks[evaloracle.Rank(c)], suits[evaloracle.Suit(c)])
	}
	return out
}

func must7(ss ...string) [7]evaloracle.Card {
	if len(ss) != 7 {
		panic("need 7 cards")
	}
	var cs [7]evaloracle.Card
	for i, s := range ss {
		cs[i] = parse(s)
	}
	return cs
}

func parse(s string) evaloracle.Card {
	if len(s) != 2 {
		panic(s)
	}
	r, su := -1, -1
	for i, ch := range ranks {
		if ch == s[0] {
			r = i
			break
		}
	}
	for i, ch := range suits {
		if ch == s[1] {
			su = i
			break
		}
	}
	if r < 0 || su < 0 {
		panic(s)
	}
	return evaloracle.Card(r + 13*su)
}
