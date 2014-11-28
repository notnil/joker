hand
========

Poker hand evaluation and ranking written in go (golang)

To install run:

```
go get github.com/SyntropyDev/joker/hand
```

```go
package main

import (
	"fmt"

	"github.com/SyntropyDev/joker/hand"
)

func main() {
	deck := hand.NewDeck()
	h1 := hand.New(deck.PopMulti(5))
	h2 := hand.New(deck.PopMulti(5))

	fmt.Println(h1)
	fmt.Println(h2)

	hands := hand.Sort(hand.SortingHigh, h1, h2)
	fmt.Println("Winner is:", hands[0].Cards())
}

```
