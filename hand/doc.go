//go:generate stringer -type=Ranking,Sorting,Ordering -output=stringer_autogen.go

/*
Package hand implements poker hand evaluation and ranking.

To install run:

	go get github.com/loganjspears/joker/hand

Example usage:

	package main

	import (
		"fmt"

		"github.com/loganjspears/joker/hand"
	)

	func main() {
		deck := hand.NewDealer().Deck()
		h1 := hand.New(deck.PopMulti(5))
		h2 := hand.New(deck.PopMulti(5))

		fmt.Println(h1)
		fmt.Println(h2)

		hands := hand.Sort(hand.SortingHigh, h1, h2)
		fmt.Println("Winner is:", hands[0].Cards())
	}
*/
package hand
