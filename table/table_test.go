package table_test

import (
	"encoding/json"
	"testing"

	"github.com/SyntropyDev/joker/hand"
	"github.com/SyntropyDev/joker/table"
)

func register() {
	table.RegisterPlayer(Player("-", []PlayerAction{}))
}

type PlayerAction struct {
	Action table.Action
	Chips  int
}

func Player(id string, actions []PlayerAction) *TestPlayer {
	return &TestPlayer{id: id, actions: actions, index: 0}
}

type TestPlayer struct {
	id      string
	actions []PlayerAction
	index   int
}

func (p *TestPlayer) Check() {
	p.actions = append(p.actions, PlayerAction{table.Check, 0})
}

func (p *TestPlayer) Call() {
	p.actions = append(p.actions, PlayerAction{table.Call, 0})
}

func (p *TestPlayer) Fold() {
	p.actions = append(p.actions, PlayerAction{table.Fold, 0})
}

func (p *TestPlayer) Bet(amount int) {
	p.actions = append(p.actions, PlayerAction{table.Bet, amount})
}

func (p *TestPlayer) Raise(amount int) {
	p.actions = append(p.actions, PlayerAction{table.Raise, amount})
}

func (p *TestPlayer) ID() string {
	return p.id
}

func (p *TestPlayer) FromID(id string) (table.Player, error) {
	return Player(id, []PlayerAction{}), nil
}

func (p *TestPlayer) Action() (a table.Action, chips int) {
	if p.index >= len(p.actions) {
		panic("player " + p.id + " doesn't have enough actions")
	}
	a = p.actions[p.index].Action
	chips = p.actions[p.index].Chips
	p.index++
	return
}

func TestToAndFronJSON(t *testing.T) {
	t.Parallel()
	register()

	// create table
	opts := table.Config{
		Game: table.Holdem,
		Stakes: table.Stakes{
			SmallBet: 1,
			BigBet:   2,
			Ante:     0,
		},
		NumOfSeats: 6,
		Limit:      table.NoLimit,
	}
	p1 := Player("1", []PlayerAction{})
	tbl := table.New(opts, hand.NewDealer())
	if err := tbl.Sit(p1, 0, 100); err != nil {
		t.Fatal(err)
	}

	// marshal into json
	b, err := json.Marshal(tbl)
	if err != nil {
		t.Fatal(err)
	}
	// unmarshal from json
	tblCopy := &table.Table{}
	if err := json.Unmarshal(b, tblCopy); err != nil {
		t.Fatal(err)
	}

	// marshal back to view
	b, err = json.Marshal(tblCopy)
	if err != nil {
		t.Fatal(err)
	}

	if len(tblCopy.Players()) != 1 {
		t.Fatal("players didn't deserialize correctly")
	}

}

func TestSeating(t *testing.T) {
	t.Parallel()

	opts := table.Config{
		Game: table.Holdem,
		Stakes: table.Stakes{
			SmallBet: 1,
			BigBet:   2,
			Ante:     0,
		},
		NumOfSeats: 6,
	}

	p1 := Player("1", []PlayerAction{})
	p1Dup := Player("1", []PlayerAction{})
	p2 := Player("2", []PlayerAction{})

	tbl := table.New(opts, hand.NewDealer())

	// sit player 1
	if err := tbl.Sit(p1, 0, 100); err != nil {
		t.Fatal(err)
	}

	// can't sit dup player 1
	err := tbl.Sit(p1Dup, 1, 100)
	if _, ok := err.(*table.Error); !ok {
		t.Fatal("should already be seated")
	}

	// can't sit player 2 in invalid seat
	if err := tbl.Sit(p2, 6, 100); err == nil {
		t.Fatal("can't sit in invalid seat")
	}

	// can't sit player 2 in occupied seat
	if err := tbl.Sit(p2, 0, 100); err == nil {
		t.Fatal("can't sit in occupied seat")
	}
}

func TestRaises(t *testing.T) {
	t.Parallel()

	opts := table.Config{
		Game: table.Holdem,
		Stakes: table.Stakes{
			SmallBet: 1,
			BigBet:   2,
			Ante:     0,
		},
		NumOfSeats: 6,
	}

	p1 := Player("1", []PlayerAction{})
	p2 := Player("2", []PlayerAction{})
	p3 := Player("3", []PlayerAction{})
	p4 := Player("4", []PlayerAction{})

	tbl := table.New(opts, hand.NewDealer())

	if err := tbl.Sit(p1, 0, 50); err != nil {
		t.Fatal(err)
	}
	if err := tbl.Sit(p2, 1, 100); err != nil {
		t.Fatal(err)
	}
	if err := tbl.Sit(p3, 2, 52); err != nil {
		t.Fatal(err)
	}
	if err := tbl.Sit(p4, 3, 60); err != nil {
		t.Fatal(err)
	}

	// preflop
	p1.Call()
	p2.Call()
	p3.Call()
	p4.Check()

	// flop
	p3.Check()
	p4.Check()
	p1.Bet(48)
	p2.Call()
	p3.Raise(2)
	p4.Raise(8)

	for i := 0; i < 12; i++ {
		if _, _, err := tbl.Next(); err != nil {
			t.Fatal(err)
		}
	}

	if tbl.Action() != 1 {
		t.Fatal("action should be on player 2")
	}

	players := tbl.Players()
	if players[1].CanRaise() {
		t.Fatal("player 2 shouldn't be able to raise")
	}

	p2.Call()
	_, _, err := tbl.Next()
	_, _, err = tbl.Next()
	if err != nil {
		t.Fatal(err)
	}
}
