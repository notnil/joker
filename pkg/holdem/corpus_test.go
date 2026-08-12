package holdem

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type potCase struct {
	Name          string `json:"name"`
	Contributions []struct {
		Seat   int  `json:"seat"`
		Chips  int  `json:"chips"`
		Folded bool `json:"folded"`
	} `json:"contributions"`
	Expect []SidePot `json:"expect"`
}

func TestPotCorpus(t *testing.T) {
	var cases []potCase
	loadJSON(t, filepath.Join("holdem", "pots.json"), &cases)
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			p := NewPot()
			for _, c := range tc.Contributions {
				p.Add(c.Seat, c.Chips)
				if c.Folded {
					p.Fold(c.Seat)
				}
			}
			got := p.SidePots()
			if len(got) != len(tc.Expect) {
				t.Fatalf("pots = %+v, want %+v", got, tc.Expect)
			}
			for i, sp := range tc.Expect {
				if got[i].Amount != sp.Amount || !sameSeats(got[i].Eligible, sp.Eligible) {
					t.Fatalf("pot %d = %+v, want %+v", i, got[i], sp)
				}
			}
		})
	}
}

type handCase struct {
	Name   string `json:"name"`
	Seats  int    `json:"seats"`
	Button int    `json:"button"`
	Stakes struct {
		SB   int `json:"sb"`
		BB   int `json:"bb"`
		Ante int `json:"ante"`
	} `json:"stakes"`
	Stacks  []int    `json:"stacks"`
	Deck    []string `json:"deck"`
	Actions []struct {
		Type  string `json:"type"`
		Chips int    `json:"chips"`
	} `json:"actions"`
	ExpectStacks []int `json:"expectStacks"`
}

func TestHandCorpus(t *testing.T) {
	var cases []handCase
	loadJSON(t, filepath.Join("holdem", "hands.json"), &cases)
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			tbl, err := NewTable(Config{
				Seats: tc.Seats,
				Stakes: Stakes{
					SmallBlind: tc.Stakes.SB,
					BigBlind:   tc.Stakes.BB,
					Ante:       tc.Stakes.Ante,
				},
			}, dealerOf(tc.Deck...))
			if err != nil {
				t.Fatal(err)
			}
			sitN(t, tbl, tc.Stacks...)
			if err := tbl.SetButton(tc.Button); err != nil {
				t.Fatal(err)
			}
			h, err := tbl.StartHand()
			if err != nil {
				t.Fatal(err)
			}
			for i, a := range tc.Actions {
				act, err := parseCorpusAction(a.Type, a.Chips)
				if err != nil {
					t.Fatalf("action %d: %v", i, err)
				}
				if err := h.Act(act); err != nil {
					t.Fatalf("action %d %v: %v (legal %v seat %d)", i, act, err, h.LegalActions(), h.ToAct())
				}
			}
			if !h.Done() {
				t.Fatal("hand not done")
			}
			if err := tbl.ApplyResults(h); err != nil {
				t.Fatal(err)
			}
			got := stacksOf(tbl)
			for i, want := range tc.ExpectStacks {
				if got[i] != want {
					t.Fatalf("stacks = %v, want %v", got, tc.ExpectStacks)
				}
			}
		})
	}
}

func parseCorpusAction(typ string, chips int) (Action, error) {
	switch typ {
	case "fold":
		return FoldAction(), nil
	case "check":
		return CheckAction(), nil
	case "call":
		return CallAction(), nil
	case "bet":
		return BetAction(chips), nil
	case "raise":
		return RaiseTo(chips), nil
	case "allin":
		return AllInAction(), nil
	default:
		return Action{}, ErrInvalidAction
	}
}

func loadJSON(t *testing.T, rel string, dest any) {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "testdata", rel)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, dest); err != nil {
		t.Fatal(err)
	}
}
