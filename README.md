# joker

Poker library in Go: hand evaluation, No-Limit Hold'em, sit-and-gos, and a browser cash table.

Requires **Go 1.26**.

```
go get github.com/notnil/joker@master
```

## Packages

| Package | Role |
| --- | --- |
| [`pkg/hand`](pkg/hand) | Cards, decks, and 5–7 card hand ranking |
| [`pkg/holdem`](pkg/holdem) | No-Limit Hold'em table, pots, and legal actions |
| [`pkg/bot`](pkg/bot) | Heuristic players (Nit, Calling Station, LAG) |
| [`pkg/play`](pkg/play) | 9-max cash session: one human seat vs bots |
| [`pkg/tournament`](pkg/tournament) | Single-table SNGs and ICM equity |
| [`pkg/jokertest`](pkg/jokertest) | Fixed decks for tests |

## Hands

```go
package main

import (
	"fmt"

	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/jokertest"
)

func main() {
	cards := jokertest.Cards("Ks", "Qs", "Js", "As", "9d")
	h := hand.New(cards)
	fmt.Println(h)            // high card ace high
	fmt.Println(h.Ranking())  // HighCard
}
```

`hand.New` selects the best five-card poker hand from the cards you pass (high or low, with optional config for games that ignore flushes or straights).

## No-Limit Hold'em

```go
package main

import (
	"log"
	"math/rand"

	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/holdem"
)

func main() {
	tbl, err := holdem.NewTable(holdem.Config{
		Seats:  6,
		Stakes: holdem.Stakes{SmallBlind: 50, BigBlind: 100},
	}, hand.NewDealer(rand.New(rand.NewSource(1))))
	if err != nil {
		log.Fatal(err)
	}
	_ = tbl.Sit(0, "Alice", 10000)
	_ = tbl.Sit(1, "Bob", 10000)

	h, err := tbl.StartHand()
	if err != nil {
		log.Fatal(err)
	}
	if err := h.Act(holdem.FoldAction()); err != nil {
		log.Fatal(err)
	}
	if h.Done() {
		_ = tbl.ApplyResults(h)
	}
}
```

`Hand.LegalActions()` lists what the player to act may do. Raises are **raise-to** (total chips in for the street). `Hand.View(hero)` is a JSON snapshot that hides other players' hole cards until showdown.

## Cash table in the browser

A 9-max cash game (50/100, 10k stacks) against eight bots:

```
make -C cmd/wasm
go run ./cmd/server
```

Open [http://localhost:8080](http://localhost:8080). HTML and CSS pick up on refresh; after rebuilding WASM, hard-refresh the page.

`pkg/play.NewCash` is the same session without a browser.

## Tournaments

`tournament.NewSNG` runs a single-table sit-and-go with a blind schedule and payouts. `tournament.ICM(stacks, payouts)` returns Malmuth–Harville chip equities as integers that sum to the prize pool.

## Tests

```
go test ./...
```
