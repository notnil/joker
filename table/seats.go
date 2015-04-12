package table

type Seats struct {
	players  map[int]*PlayerState
	minBuyin int
	maxBuyin int
}

// Len returns the total number of seats
func (s *Seats) Len() int {
	return len(s.players)
}

func (s *Seats) Players() map[int]*PlayerState {
	return s.players
}

// Sit sits the player with the given amount of chips. An error is
// return if the seat is invalid, the player is already seated, the
// seat is already occupied, or the chips are outside the valid buy
// in amounts.
func (s *Seats) Sit(p Player, seat, chips int) error {
	if !s.validSeat(seat) {
		return ErrInvalidSeat
	} else if s.isSeated(p) {
		return ErrAlreadySeated
	} else if _, occupied := s.players[seat]; occupied {
		return ErrSeatOccupied
	}

	if chips < s.minBuyin || chips > s.maxBuyin {
		return ErrInvalidBuyin
	}

	s.players[seat] = &PlayerState{
		player:    p,
		holeCards: []*HoleCard{},
		chips:     chips,
	}
	return nil
}

// Stand unseats the player.  If the player isn't seated the command
// is ignored.
func (s *Seats) Stand(p Player) {
	for seat, pl := range s.players {
		if pl.player.ID() == p.ID() {
			delete(s.players, seat)
			return
		}
	}
}

func newSeats(n, minBuyin, maxBuyin int) *Seats {
	players := map[int]*PlayerState{}
	for i := 0; i < n; i++ {
		players[i] = nil
	}
	return &Seats{
		players: players,
	}
}

func (s *Seats) next(seat int, playing bool) int {
	count := 0
	length := s.Len()
	seat = seat % length
	for count < length {
		p, ok := s.players[seat]
		if ok && (!playing || (!p.out && !p.allin && !p.acted)) {
			return seat
		}
		count++
		seat = (seat + 1) % length
	}
	return -1
}

func (s *Seats) relativePos(button, seat int) int {
	current := button
	count := 0
	for {
		if current == seat {
			break
		}
		current = s.next(current+1, false)
		count++
	}
	return count
}

func (s *Seats) validSeat(seat int) bool {
	return seat >= 0 && seat < s.Len()
}

func (s *Seats) isSeated(p Player) bool {
	for _, pl := range s.players {
		if p.ID() == pl.player.ID() {
			return true
		}
	}
	return false
}
