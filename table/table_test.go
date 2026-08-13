package table

import "testing"

func TestSeatingAndFind(t *testing.T) {
    tb := New(Config{MaxSeats: 6})

    // Autoseat two players
    s1, err := tb.AutoSeat(Player{ID: "p1", Name: "Alice"})
    if err != nil {
        t.Fatalf("autoseat p1: %v", err)
    }
    s2, err := tb.AutoSeat(Player{ID: "p2", Name: "Bob"})
    if err != nil {
        t.Fatalf("autoseat p2: %v", err)
    }
    if s1 == s2 || s1 == NoSeat || s2 == NoSeat {
        t.Fatalf("unexpected seats: s1=%v s2=%v", s1, s2)
    }
    if tb.NumSeated() != 2 {
        t.Fatalf("expected 2 seated, got %d", tb.NumSeated())
    }

    // Find players
    if seat, ok := tb.FindPlayer("p1"); !ok || seat != s1 {
        t.Fatalf("expected find p1 at %v, got %v %v", s1, seat, ok)
    }
    if seat, ok := tb.FindPlayer("zzz"); ok || seat != NoSeat {
        t.Fatalf("expected not find zzz, got %v %v", seat, ok)
    }
}

func TestSeatAtAndLeave(t *testing.T) {
    tb := New(Config{MaxSeats: 3})
    if err := tb.SeatAt(SeatIndex(1), Player{ID: "p1"}); err != nil {
        t.Fatalf("seat at 1: %v", err)
    }
    if err := tb.SeatAt(SeatIndex(1), Player{ID: "p2"}); err == nil {
        t.Fatalf("expected seat occupied error")
    }
    if err := tb.Leave(SeatIndex(1)); err != nil {
        t.Fatalf("leave 1: %v", err)
    }
    if err := tb.Leave(SeatIndex(1)); err == nil {
        t.Fatalf("expected seat empty error")
    }
}

func TestActionFlow(t *testing.T) {
    tb := New(Config{MaxSeats: 4})
    s1, _ := tb.AutoSeat(Player{ID: "p1"})
    s2, _ := tb.AutoSeat(Player{ID: "p2"})
    s3, _ := tb.AutoSeat(Player{ID: "p3"})

    if tb.Dealer() != s1 {
        t.Fatalf("expected dealer to start at first seated: %v got %v", s1, tb.Dealer())
    }
    if err := tb.StartHand(); err != nil {
        t.Fatalf("start hand: %v", err)
    }
    if tb.ToAct() != s2 {
        t.Fatalf("expected to act to be left of dealer: %v got %v", s2, tb.ToAct())
    }
    next, err := tb.NextToAct()
    if err != nil || next != s3 {
        t.Fatalf("expected next to act s3: got %v err %v", next, err)
    }
    adv, err := tb.AdvanceAction()
    if err != nil || adv != s3 || tb.ToAct() != s3 {
        t.Fatalf("advance to s3 failed: adv=%v err=%v toAct=%v", adv, err, tb.ToAct())
    }
    if err := tb.AdvanceDealer(); err != nil {
        t.Fatalf("advance dealer: %v", err)
    }
    if tb.Dealer() != s2 {
        t.Fatalf("expected dealer to rotate to s2, got %v", tb.Dealer())
    }
    tb.EndHand()
    if tb.ToAct() != NoSeat {
        t.Fatalf("expected toAct reset after EndHand")
    }
}

