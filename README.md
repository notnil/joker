joker
========
[![Build Status](https://drone.io/github.com/loganjspears/joker/status.png)](https://drone.io/github.com/loganjspears/joker/latest)


The goal of Joker is to be an open source and fully functioning poker backend in go.  Joker will attempt to support roughly the same feature set as Full Tilt Poker.  

## cmd

cmd holds the executable portions of Joker.  This currently includes a comand line client demoing table functionality.

## hand

The [hand package](http://www.godoc.org/github.com/loganjspears/joker/hand) is responsible for poker hand evaluation.  hand is also home to card and deck implementations.  

## jokertest

The [jokertest package](http://www.godoc.org/github.com/loganjspears/joker/jokertest) provides convience methods for testing in the other packages.  For example, jokertest's Dealer produces a Deck with a prearranged series of cards instead of ones in random order.  

## pot

The [pot package](http://www.godoc.org/github.com/loganjspears/joker/pot) tracks contributions from players and awards players with winning hands.  It supports hi/lo split pots.  (pot might eventually get merged into table)

## table

The [table package](http://www.godoc.org/github.com/loganjspears/joker/table) provides a table engine to run a poker table.  Turn managment, player action requests, dealing, forced bets, etc are in this package.  An example of a working table is available in th cmd section.  
## util

util is a place for code shared by multiple packages, but otherwise wouldn't be exported.  Might be converted to internal package in go 1.5.
