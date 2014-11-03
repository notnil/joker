/*
Package joker implements poker hand evaluation and ranking.

To install run:

	go get github.com/SyntropyDev/joker

Example usage:

	import (
		"fmt"
		"sort"

		"github.com/SyntropyDev/joker"
	)

	func main() {
		deck := joker.NewDeck()
		h1 := joker.NewHand(deck.PopMulti(5))
		h2 := joker.NewHand(deck.PopMulti(5))

		fmt.Println(h1)
		fmt.Println(h2)

		hands := []*joker.Hand{h1, h2}
		sort.Sort(joker.ByHighHand(hands))

		fmt.Println("Winner is:", hands[1].Cards())
	}
*/
package joker
