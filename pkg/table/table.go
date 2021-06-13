package table

import (
	"encoding/json"
	"errors"

	"github.com/notnil/joker/hand"
)

var (
	ErrInvalidSeatCount = errors.New("tables must have between 2 and 10 seats")
	ErrInvalidSeat      = errors.New("invalid seat")
	ErrSeatOccupied     = errors.New("seat occupied")
	ErrInvalidBuyIn     = errors.New("invalid buyin")
)

type Variant int

const (
	TexasHoldem Variant = iota
	OmahaHi
)

type Limit int

const (
	NoLimit Limit = iota
	PotLimit
)

type Stakes struct {
	BigBlind   int `json:"bigBlind"`
	SmallBlind int `json:"smallBlind"`
	Ante       int `json:"ante"`
}

type Config struct {
	Size     int     `json:"size"`
	BuyInMin int     `json:"buyInMin"`
	BuyInMax int     `json:"buyInMax"`
	Variant  Variant `json:"variant"`
	Stakes   Stakes  `json:"stakes"`
	Limit    Limit   `json:"limit"`
}

type Player struct {
	ID    string
	Chips int
}

type Table struct {
	seats  map[int]*Player
	config Config
	dealer hand.Dealer
	button int
}

func New(c Config, seats map[int]*Player, d hand.Dealer) (*Table, error) {
	if c.Size < 2 || c.Size > 10 {
		return nil, ErrInvalidSeatCount
	}
	t := &Table{seats: map[int]*Player{}, config: c, button: 0, dealer: d}
	for k, v := range seats {
		if err := t.Sit(k, v); err != nil {
			return nil, err
		}
	}
	t.button = t.Next(t.button)
	return t, nil
}

func (t *Table) Sit(i int, p *Player) error {
	if i < 0 || i >= t.config.Size {
		return ErrInvalidSeat
	}
	_, occupied := t.seats[i]
	if occupied {
		return ErrSeatOccupied
	}
	if p.Chips < t.config.BuyInMin || p.Chips > t.config.BuyInMax {
		return ErrInvalidBuyIn
	}
	t.seats[i] = p
	return nil
}

func (t *Table) StandUp(i int) error {
	if i < 0 || i >= t.config.Size {
		return ErrInvalidSeat
	}
	delete(t.seats, i)
	return nil
}

func (t *Table) Player(i int) *Player {
	return t.seats[i]
}

func (t *Table) Players() map[int]*Player {
	cp := map[int]*Player{}
	for k, v := range t.seats {
		cp[k] = v
	}
	return cp
}

func (t *Table) Config() Config {
	return t.config
}

func (t *Table) PlayerCount() int {
	return len(t.seats)
}

func (t *Table) Next(i int) int {
	if i < 0 || i >= t.config.Size {
		return -1
	}
	for j := 1; j <= t.config.Size; j++ {
		next := (i + j) % t.config.Size
		if _, occupied := t.seats[next]; occupied {
			return next
		}
	}
	return -1
}

func (t *Table) NewHand() *Hand {
	seats := map[int]*PlayerInHand{}
	for seat, player := range t.seats {
		seats[seat] = &PlayerInHand{
			ID:    player.ID,
			Chips: player.Chips,
			Seat:  seat,
		}
	}
	h := &Hand{
		Table: t,
		Pot:   NewPot(nil),
		Deck:  t.dealer.Deck(),
		Seats: seats,
	}
	h.setupRound()
	return h
}

func (t *Table) Update(h *Hand) {
	t.button = t.Next(t.button)
	for seat, player := range h.Seats {
		t.seats[seat].Chips = player.Chips
	}
	for seat, results := range h.Results {
		for _, result := range results {
			t.seats[seat].Chips += result.Chips
		}
	}
}

type tableJSON struct {
	Seats  map[int]*Player `json:"seats"`
	Config Config          `json:"config"`
	Button int             `json:"button"`
}

func (t *Table) MarshalJSON() ([]byte, error) {
	js := &tableJSON{
		Seats:  t.seats,
		Config: t.config,
		Button: t.button,
	}
	return json.Marshal(js)
}
