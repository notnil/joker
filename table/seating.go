package table

// Public seating helpers to provide higher-level operations or utilities.

// FirstEmptySeat returns the first empty SeatIndex or NoSeat if none.
func (t *tableImpl) FirstEmptySeat() SeatIndex {
    for i := 0; i < t.cfg.MaxSeats; i++ {
        if t.state.Seats[i].Status == SeatEmpty {
            return SeatIndex(i)
        }
    }
    return NoSeat
}

