package holdem

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestViewHidesOtherHoles(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf(
		"Ah", "Ad", "Kc", "Kd", "2s", "3s",
	))
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 1000, 1000, 1000)
	_ = tbl.SetButton(0)
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	v := h.View(0)
	if v.Hero != 0 {
		t.Fatalf("hero = %d", v.Hero)
	}
	seen := map[int]int{}
	for _, p := range v.Players {
		seen[p.Seat] = len(p.Hole)
	}
	if seen[0] != 2 {
		t.Fatalf("hero should see own hole, got %d", seen[0])
	}
	if seen[1] != 0 || seen[2] != 0 {
		t.Fatalf("opponents leaked holes: %+v", v.Players)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	raw := string(b)
	if strings.Contains(raw, "A♥") || strings.Contains(raw, "K♣") {
		t.Fatalf("json leaked opponent cards: %s", raw)
	}
}

func TestViewShowdownRevealsContesting(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 2, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf())
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 1000, 1000)
	_ = tbl.SetButton(0)
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	mustAct(t, h, FoldAction())
	v := h.View(0)
	if !v.Done {
		t.Fatal("expected done")
	}
	for _, p := range v.Players {
		if p.Seat == 1 && len(p.Hole) == 0 {
			t.Fatal("winner hole should be visible after fold-win? fold-win contesting is BB only")
		}
	}
}

func TestViewSidePots(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf(
		"Ah", "Ad",
		"Kc", "Kd",
		"2s", "3s",
		"2h", "7c", "8d", "9s", "Jc",
	))
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 1000, 300, 500)
	_ = tbl.SetButton(0)
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	mustAct(t, h, RaiseTo(1000))
	mustAct(t, h, AllInAction())
	mustAct(t, h, AllInAction())
	v := h.View(0)
	if len(v.Pots) < 2 {
		t.Fatalf("want side pots, got %+v", v.Pots)
	}
	total := 0
	for _, p := range v.Pots {
		total += p.Amount
	}
	if total != v.Pot {
		t.Fatalf("pot parts %d != total %d", total, v.Pot)
	}
}

func TestParseAction(t *testing.T) {
	a, err := ParseAction("RAISE", 300)
	if err != nil || a.Type != Raise || a.Chips != 300 {
		t.Fatalf("got %+v %v", a, err)
	}
	if _, err := ParseAction("nope", 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestTableSetChips(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 2, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf())
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 100, 100)
	if err := tbl.SetChips(0, 10000); err != nil {
		t.Fatal(err)
	}
	if tbl.Player(0).Chips != 10000 {
		t.Fatal("rebuy failed")
	}
}
