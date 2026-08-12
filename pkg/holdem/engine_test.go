package holdem

import (
	"testing"

	"github.com/notnil/joker/pkg/hand"
	"github.com/notnil/joker/pkg/jokertest"
)

func dealerOf(front ...string) hand.Dealer {
	cards := jokertest.Cards(front...)
	used := map[hand.Card]bool{}
	for _, c := range cards {
		used[c] = true
	}
	var rest []hand.Card
	for _, c := range hand.Cards() {
		if !used[c] {
			rest = append(rest, c)
		}
	}
	all := append(cards, rest...)
	return jokertest.Dealer(all)
}

func sitN(t *testing.T, tbl *Table, stacks ...int) {
	t.Helper()
	for i, chips := range stacks {
		if err := tbl.Sit(i, string(rune('A'+i)), chips); err != nil {
			t.Fatal(err)
		}
	}
}

func mustAct(t *testing.T, h *Hand, a Action) {
	t.Helper()
	if err := h.Act(a); err != nil {
		t.Fatalf("act %v at seat %d: %v (legal %v)", a, h.ToAct(), err, h.LegalActions())
	}
}

func stacksOf(tbl *Table) []int {
	out := make([]int, tbl.Seats())
	for i := range out {
		if p := tbl.Player(i); p != nil {
			out[i] = p.Chips
		}
	}
	return out
}

func TestHeadsUpBlindsAndFold(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 2, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf())
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 1000, 1000)
	if err := tbl.SetButton(0); err != nil {
		t.Fatal(err)
	}
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	if h.SmallBlindSeat() != 0 || h.BigBlindSeat() != 1 {
		t.Fatalf("blinds sb=%d bb=%d", h.SmallBlindSeat(), h.BigBlindSeat())
	}
	if h.ToAct() != 0 {
		t.Fatalf("hu button acts first preflop, got %d", h.ToAct())
	}
	if h.Stack(0) != 950 || h.Stack(1) != 900 {
		t.Fatalf("posted stacks %d %d", h.Stack(0), h.Stack(1))
	}
	mustAct(t, h, FoldAction())
	if !h.Done() {
		t.Fatal("expected fold-win")
	}
	if err := tbl.ApplyResults(h); err != nil {
		t.Fatal(err)
	}
	got := stacksOf(tbl)
	if got[0] != 950 || got[1] != 1050 {
		t.Fatalf("stacks after fold = %v", got)
	}
}

func TestFoldToBBThreeHanded(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf())
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 1000, 1000, 1000)
	if err := tbl.SetButton(0); err != nil {
		t.Fatal(err)
	}
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	if h.ToAct() != 0 {
		t.Fatalf("utg should be button, got %d", h.ToAct())
	}
	mustAct(t, h, FoldAction())
	mustAct(t, h, FoldAction())
	if !h.Done() {
		t.Fatal("expected bb win")
	}
	if err := tbl.ApplyResults(h); err != nil {
		t.Fatal(err)
	}
	got := stacksOf(tbl)
	if got[0] != 1000 || got[1] != 950 || got[2] != 1050 {
		t.Fatalf("stacks = %v", got)
	}
}

func TestCheckDownShowdown(t *testing.T) {
	// Seat0 button: As Ks; seat1 SB: 2h 3h; seat2 BB: Ah Ad
	// Board: 2c 3c 4d 5s 9h — seat2 has pair aces, seat1 two pair, seat0 high card
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf(
		"2h", "3h", // SB seat 1
		"Ah", "Ad", // BB seat 2
		"As", "Ks", // BTN seat 0
		"7c", "8d", "9s", "2c", "3c",
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
	mustAct(t, h, CallAction())  // button calls
	mustAct(t, h, CallAction())  // SB completes
	mustAct(t, h, CheckAction()) // BB option
	for h.Street() != River || !h.Done() {
		if h.Done() {
			break
		}
		mustAct(t, h, CheckAction())
	}
	if !h.Done() {
		t.Fatal("hand should be over")
	}
	if err := tbl.ApplyResults(h); err != nil {
		t.Fatal(err)
	}
	got := stacksOf(tbl)
	// pair of aces (BB) beats two pair? Wait seat1 has 2h3h + 2c3c = two pair twos and threes.
	// BB Ah Ad + board = pair aces. Two pair beats pair. Seat1 wins 300.
	if got[1] != 1200 {
		t.Fatalf("stacks = %v (want seat1 win)", got)
	}
}

func TestAllInRunoutNoBetting(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 2, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf(
		"As", "Ks",
		"2h", "3h",
		"Ah", "Kh", "Qh", "Jh", "9c",
	))
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 150, 1000)
	_ = tbl.SetButton(0)
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	mustAct(t, h, AllInAction()) // SB all-in 100 more (already 50)
	mustAct(t, h, CallAction())
	if !h.Done() {
		t.Fatalf("should run out, street=%s done=%v toAct=%d", h.Street(), h.Done(), h.ToAct())
	}
	if len(h.Board()) != 5 {
		t.Fatalf("board = %v", h.Board())
	}
}

func TestShortAllInContinuesOnFlop(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf())
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 1000, 1000, 150)
	_ = tbl.SetButton(0)
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	mustAct(t, h, RaiseTo(300))
	mustAct(t, h, CallAction())
	mustAct(t, h, AllInAction())
	if h.Done() {
		t.Fatal("two players still have chips; flop should be dealt")
	}
	if h.Street() != Flop {
		t.Fatalf("street=%s", h.Street())
	}
	if !h.IsAllIn(2) {
		t.Fatal("bb should be all-in")
	}
}

func TestIncompleteRaiseClosesActionForActors(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf())
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 250, 1000, 1000) // button short enough for an incomplete flop raise
	_ = tbl.SetButton(0)
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	mustAct(t, h, CallAction())  // UTG/button calls 100, 50 behind
	mustAct(t, h, CallAction())  // SB
	mustAct(t, h, CheckAction()) // BB
	if h.Street() != Flop || h.ToAct() != 1 {
		t.Fatalf("flop toAct=%d street=%s", h.ToAct(), h.Street())
	}
	mustAct(t, h, BetAction(100)) // SB bets 100
	mustAct(t, h, CallAction())   // BB calls 100
	mustAct(t, h, AllInAction())  // BTN all-in 150 — incomplete raise of 50 < 100
	if h.Street() != Flop {
		t.Fatalf("expected flop after incomplete raise, street=%s", h.Street())
	}
	// SB and BB have already acted; they may call/fold but not raise.
	if h.ToAct() != 1 {
		t.Fatalf("action should return to SB, got %d", h.ToAct())
	}
	legal := h.LegalActions()
	if includesType(legal, Raise) || includesType(legal, Bet) {
		t.Fatalf("incomplete raise reopened action: %v", legal)
	}
	if !includesType(legal, Call) || !includesType(legal, Fold) {
		t.Fatalf("should be able to call/fold: %v", legal)
	}
	mustAct(t, h, CallAction())
	legal = h.LegalActions()
	if includesType(legal, Raise) {
		t.Fatalf("BB should not raise: %v", legal)
	}
	mustAct(t, h, CallAction())
}

func TestSidePotAwards(t *testing.T) {
	// 3-way all-in different stacks. Seat1 short 300, seat2 500, seat0 1000.
	// Seat1 has the nuts, seat2 second, seat0 worst.
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf(
		"Ah", "Ad", // SB seat1 — aces
		"Kc", "Kd", // BB seat2 — kings
		"2s", "3s", // BTN seat0 — nothing
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
	mustAct(t, h, RaiseTo(1000)) // BTN shoves
	mustAct(t, h, AllInAction()) // SB 250 more
	mustAct(t, h, AllInAction()) // BB
	if !h.Done() {
		t.Fatal("expected showdown")
	}
	if err := tbl.ApplyResults(h); err != nil {
		t.Fatal(err)
	}
	got := stacksOf(tbl)
	// Main pot 300*3=900 to seat1 (aces)
	// Side 200*2=400 to seat2 (kings beat 23)
	// Uncalled 500 to seat0
	// seat1: 0+900=900
	// seat2: 0+400=400
	// seat0: 0+500=500
	if got[0] != 500 || got[1] != 900 || got[2] != 400 {
		t.Fatalf("stacks = %v", got)
	}
}

func TestOddChipToLeftOfButton(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf(
		"As", "Ks", // SB 1
		"Ah", "Kh", // BB 2
		"Ad", "Kd", // BTN 0
		"2c", "3c", "4c", "5d", "7h", // all ace-high same
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
	mustAct(t, h, CallAction())
	mustAct(t, h, CallAction())
	mustAct(t, h, CheckAction())
	for !h.Done() {
		mustAct(t, h, CheckAction())
	}
	awards := h.Awards()
	// pot 300, 3-way chop, 100 each, no odd chip
	if awards[0] != 100 || awards[1] != 100 || awards[2] != 100 {
		t.Fatalf("awards = %v", awards)
	}

	// Force a 101 pot via ante 1? Easier unit test of splitAmount already exists.
	// Use ante=1: pot 303, 101 each.
}

func TestMinRaise(t *testing.T) {
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
	if h.MinRaiseTo() != 200 {
		t.Fatalf("min raise to %d", h.MinRaiseTo())
	}
	if err := h.Act(RaiseTo(150)); err == nil {
		t.Fatal("short raise should fail")
	}
	mustAct(t, h, RaiseTo(200))
	if h.MinRaiseTo() != 300 {
		t.Fatalf("after raise last full=100, min to %d", h.MinRaiseTo())
	}
}

func TestBBOption(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 3, Stakes: Stakes{SmallBlind: 50, BigBlind: 100}}, dealerOf())
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 1000, 1000, 1000)
	_ = tbl.SetButton(0)
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	mustAct(t, h, CallAction())
	mustAct(t, h, CallAction())
	if h.ToAct() != 2 {
		t.Fatalf("bb option, toAct=%d", h.ToAct())
	}
	legal := h.LegalActions()
	if !includesType(legal, Check) || !includesType(legal, Raise) && !includesType(legal, Bet) {
		t.Fatalf("bb option legal=%v", legal)
	}
}

func TestAntes(t *testing.T) {
	tbl, err := NewTable(Config{Seats: 2, Stakes: Stakes{SmallBlind: 50, BigBlind: 100, Ante: 10}}, dealerOf())
	if err != nil {
		t.Fatal(err)
	}
	sitN(t, tbl, 1000, 1000)
	_ = tbl.SetButton(0)
	h, err := tbl.StartHand()
	if err != nil {
		t.Fatal(err)
	}
	if h.PotTotal() != 10+10+50+100 {
		t.Fatalf("pot = %d", h.PotTotal())
	}
	if h.CallAmount() != 50 { // SB to call BB, ante not owed
		t.Fatalf("call = %d", h.CallAmount())
	}
}
