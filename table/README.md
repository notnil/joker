# table

A small, modular poker table package for seating and action flow.

## Features

- Seating: `SeatAt`, `AutoSeat`, `Leave`, occupancy checks
- Action: `StartHand`, `AdvanceAction`, `AdvanceDealer`, `EndHand`
- Simple state snapshots

## Install

```bash
go get github.com/notnil/joker/table
```

## Usage

```go
tb := table.New(table.Config{MaxSeats: 6})
_, _ = tb.AutoSeat(table.Player{ID: "p1", Name: "Alice"})
_, _ = tb.AutoSeat(table.Player{ID: "p2", Name: "Bob"})

_ = tb.StartHand()
dealer := tb.Dealer()
toAct := tb.ToAct()
next, _ := tb.AdvanceAction()
_ = tb.AdvanceDealer()
tb.EndHand()

_ = tb.Leave(dealer)
```

