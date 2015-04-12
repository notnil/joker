package table

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/loganjspears/joker/hand"
	"github.com/loganjspears/joker/pot"
)

var (
	// ErrInvalidBuyin errors occur when a player attempts to sit at a
	// table with an invalid buyin.
	ErrInvalidBuyin = errors.New("table: player attempted sitting with invalid buyin")

	// ErrSeatOccupied errors occur when a player attempts to sit at a
	// table in a seat that is already occupied.
	ErrSeatOccupied = errors.New("table: player attempted sitting in occupied seat")

	// ErrInvalidSeat errors occur when a player attempts to sit at a
	// table in a seat that is invalid.
	ErrInvalidSeat = errors.New("table: player attempted sitting in invalid seat")

	// ErrAlreadySeated errors occur when a player attempts to sit at a
	// table at which the player is already seated.
	ErrAlreadySeated = errors.New("table: player attempted sitting when already seated")

	// ErrInsufficientPlayers errors occur when the table's Next() method
	// can't start a new hand because of insufficient players
	ErrInsufficientPlayers = errors.New("table: insufficent players for call to table's Next() method")

	// ErrInvalidBet errors occur when a player attempts to bet an invalid
	// amount.  Bets are invalid if they exceed a player's chips or fall below the
	// stakes minimum bet.  In fixed limit games the bet amount must equal the amount
	// prespecified by the limit and round.  In pot limit games the bet must be less
	// than or equal to the pot.
	ErrInvalidBet = errors.New("table: player attempted invalid bet")

	// ErrInvalidRaise errors occur when a player attempts to raise an invalid
	// amount.  Raises are invalid if the raise or reraise is lower than the previous bet
	// or raised amount unless it puts the player allin.  Raises are also invalid if they
	// exceed a player's chips. In fixed limit games the raise amount must equal the amount
	// prespecified by the limit and round.  In pot limit games the raise must be less
	// than or equal to the pot.
	ErrInvalidRaise = errors.New("table: player attempted invalid raise")

	// ErrInvalidAction errors occur when a player attempts an action that isn't
	// currently allowed.  For example a check action is invalid when faced with a raise.
	ErrInvalidAction = errors.New("table: player attempted invalid action")
)

// An Action is an action a player can take in a hand.
type Action string

const (
	// Fold discards one's hand and forfeits interest in
	// the current pot.
	Fold Action = "Fold"

	// Check is the forfeit to bet when not faced with a bet or
	// raise.
	Check Action = "Check"

	// Call is a match of a bet or raise.
	Call Action = "Call"

	// Bet is a wager that others must match to remain a contender
	// in the current pot.
	Bet Action = "Bet"

	// Raise is an increase to the original bet that others must
	// match to remain a contender in the current pot.
	Raise Action = "Raise"
)

// Stakes are the forced bet amounts for the table.
type Stakes struct {

	// SmallBet is the smaller forced bet amount.
	SmallBet int `json:"smallBet"`

	// BigBet is the bigger forced bet amount.
	BigBet int `json:"bigBet"`

	// Ante is the amount requried from each player to start the hand.
	Ante int `json:"ante"`
}

// Limit is the bet and raise limits of a poker game
type Limit string

const (
	// NoLimit has no limit and players may go "all in"
	NoLimit Limit = "NL"

	// PotLimit has the current value of the pot as the limit
	PotLimit Limit = "PL"

	// FixedLimit restricted the size of bets and raises to predefined
	// values based on the game and round.
	FixedLimit Limit = "FL"
)

// Config are the configurations for creating a table.
type Config struct {

	// Game is the game of the table.
	Game Game `json:"game"`

	// Limit is the limit of the table
	Limit Limit `json:"limit"`

	// Stakes is the stakes for the table.
	Stakes Stakes `json:"stakes"`

	// NumOfSeats is the number of seats available for the table.
	NumOfSeats int `json:"numOfSeats"`

	// MinBuyin is the minimum buyin a player can sit with
	MinBuyin int `json:"minBuyin"`

	// MaxBuyin is the maximum buyin a player can sit with
	MaxBuyin int `json:"maxBuyin"`
}

// Table represent a poker table and dealer.  A table manages the
// game state and all player interactions at the table.
type Table struct {
	opts        Config
	dealer      hand.Dealer
	deck        *hand.Deck
	button      int
	action      int
	round       int
	minRaise    int
	board       []*hand.Card
	seats       *Seats
	pot         *pot.Pot
	startedHand bool
}

// New creates a new table with the options and deck provided.  To
// start playing hands, at least two players must be seated and the
// Next() function must be called.  If the number of seats is invalid
// for the Game specified New panics.
func New(opts Config, dealer hand.Dealer) *Table {
	if int(opts.NumOfSeats) > opts.Game.get().MaxSeats() {
		format := "table: %s has a maximum of %d seats but attempted %d"
		s := fmt.Sprintf(format, opts.Game, opts.Game.get().MaxSeats(), opts.NumOfSeats)
		panic(s)
	}

	seats := newSeats(opts.NumOfSeats, opts.MinBuyin, opts.MaxBuyin)
	return &Table{
		opts:   opts,
		dealer: dealer,
		deck:   dealer.Deck(),
		board:  []*hand.Card{},
		seats:  seats,
		pot:    pot.New(int(opts.NumOfSeats)),
		action: -1,
	}
}

// Action returns the seat that the action is currently on.  If no
// seat has the action then -1 is returned.
func (t *Table) Action() int {
	return t.action
}

// Board returns the current community cards.  An empty slice is
// returned if there are no community cards or the game doesn't
// support community cards.
func (t *Table) Board() []*hand.Card {
	c := []*hand.Card{}
	return append(c, t.board...)
}

// Button returns the seat that the button is currently on.
func (t *Table) Button() int {
	return t.button
}

// CurrentPlayer returns the player the action is currently on.  If
// no player is current then it returns nil.
func (t *Table) CurrentPlayer() *PlayerState {
	return t.seats.Players()[t.Action()]
}

// Game returns the game of the table.
func (t *Table) Game() Game {
	return t.opts.Game
}

// Limit returns the limit of the table.
func (t *Table) Limit() Limit {
	return t.opts.Limit
}

// MaxRaise returns the maximum number of chips that can be bet or
// raised by the current player.  If there is no current player then
// 0 is returned.
func (t *Table) MaxRaise() int {
	player := t.CurrentPlayer()
	if isNil(player) {
		return 0
	}

	outstanding := t.Outstanding()
	chips := player.Chips()
	bettableChips := chips - outstanding

	if bettableChips <= 0 {
		return 0
	}

	if !player.CanRaise() {
		return 0
	}

	max := bettableChips
	switch t.opts.Limit {
	case PotLimit:
		max = t.pot.Chips() + outstanding
	case FixedLimit:
		max = t.game().FixedLimit(t.opts, round(t.round))
	}
	if max > bettableChips {
		max = bettableChips
	}
	return max
}

// MinRaise returns the minimum number of chips that can be bet or
// raised by the current player. If there is no current player then
// 0 is returned.
func (t *Table) MinRaise() int {
	player := t.CurrentPlayer()
	if isNil(player) {
		return 0
	}

	outstanding := t.Outstanding()
	bettableChips := player.Chips() - outstanding

	if !player.CanRaise() {
		return 0
	}

	if bettableChips < t.minRaise {
		return bettableChips
	}
	return t.minRaise
}

// NumOfSeats returns the number of seats.
func (t *Table) NumOfSeats() int {
	return int(t.opts.NumOfSeats)
}

// Outstanding returns the number of chips owed to the pot by the
// current player.  If there is no current player then 0 is returned.
func (t *Table) Outstanding() int {
	player := t.CurrentPlayer()
	if isNil(player) || player.AllIn() || player.Out() {
		return 0
	}
	return t.pot.Outstanding(t.Action())
}

// Players returns a mapping of seats to player states.
func (t *Table) Players() map[int]*PlayerState {
	return t.seats.Players()
}

// View returns a view of the table that only contains information
// privileged to the given player.
func (t *Table) View(p Player) *Table {
	players := map[int]*PlayerState{}
	for seat, player := range t.seats.Players() {
		if p.ID() == player.Player().ID() {
			players[seat] = player
			continue
		}

		players[seat] = &PlayerState{
			player:    player.player,
			holeCards: tableViewOfHoleCards(player.holeCards),
			chips:     player.chips,
			acted:     player.acted,
			out:       player.out,
			allin:     player.allin,
			canRaise:  player.canRaise,
		}
	}
	s := &Seats{
		players: players,
	}
	return &Table{
		opts:        t.opts,
		deck:        &hand.Deck{Cards: []*hand.Card{}},
		button:      t.button,
		action:      t.action,
		round:       t.round,
		minRaise:    t.minRaise,
		board:       t.board,
		pot:         t.pot,
		startedHand: t.startedHand,
		seats:       s,
	}
}

// Pot returns the current pot.
func (t *Table) Pot() *pot.Pot {
	return t.pot
}

// Round returns the current round.
func (t *Table) Round() int {
	return t.round
}

// Stakes returns the stakes.
func (t *Table) Stakes() Stakes {
	return t.opts.Stakes
}

// String returns a string useful for debugging.
func (t *Table) String() string {
	const format = "{Button: Seat %d, Current Player: %s, Round %d, Board: %s, Pot: %d}\n"
	current := "None"
	if t.action != -1 && !isNil(t.CurrentPlayer()) {
		current = t.CurrentPlayer().player.ID()
	}

	return fmt.Sprintf(format, t.button, current, t.round, t.board, t.pot.Chips())
}

// ValidActions returns the actions that can be taken by the current
// player.
func (t *Table) ValidActions() []Action {
	player := t.CurrentPlayer()
	if player.AllIn() || player.Out() {
		return []Action{}
	}

	if t.Outstanding() == 0 {
		return []Action{Check, Bet}
	}

	if !player.CanRaise() {
		return []Action{Fold, Call}
	}

	return []Action{Fold, Call, Raise}
}

// Next is the iterator function of the table.  Next updates the
// table's state while calling player's Action() method to get
// an action for each player's turn.  New hands are started
// automatically if there are two or more eligible players.  Next
// moves through each round of betting until the showdown at which
// point are paid out.  The results are returned as a map of seats
// to pot results. If the round is not a showdown then results are
// nil. err is nil unless there are insufficient players to start
// the next hand or a player's action has an error. done indicates
// that the table can not continue.
func (t *Table) Next() (results map[int][]*pot.Result, done bool, err error) {
	if !t.startedHand {
		if !t.hasNextHand() {
			return nil, true, ErrInsufficientPlayers
		}
		t.setUpHand()
		t.setUpRound()
		t.startedHand = true
		return nil, false, nil
	}

	if t.action == -1 {
		t.round++

		if t.round == t.game().NumOfRounds() {
			holeCards := cardsFromHoleCardMap(t.holeCards())
			highHands := pot.NewHands(holeCards, t.board, t.game().FormHighHand)
			lowHands := pot.NewHands(holeCards, t.board, t.game().FormLowHand)
			results = t.pot.Payout(highHands, lowHands, t.game().Sorting(), t.button)
			t.payoutResults(results)
			t.startedHand = false
			return results, false, nil
		}

		t.setUpRound()
		return nil, false, nil
	}

	current := t.CurrentPlayer()
	action, chips := current.player.Action()

	if err := t.handleAction(t.action, current, action, chips); err != nil {
		return nil, false, err
	}

	// check if only one person left
	if t.everyoneFolded() {
		for seat, player := range t.seats.Players() {
			if player.out {
				continue
			}
			results = t.pot.Take(seat)
			t.payoutResults(results)
			t.startedHand = false
			return results, false, nil
		}
	}

	t.action = t.seats.next(t.action+1, true)
	return nil, false, nil
}

type tableJSON struct {
	Options     Config                  `json:"options"`
	Deck        *hand.Deck              `json:"deck"`
	Button      int                     `json:"button"`
	Action      int                     `json:"action"`
	Round       int                     `json:"round"`
	MinRaise    int                     `json:"minRaise"`
	Board       []*hand.Card            `json:"board"`
	Players     map[string]*PlayerState `json:"players"`
	Pot         *pot.Pot                `json:"pot"`
	StartedHand bool                    `json:"startedHand"`
}

// MarshalJSON implements the json.Marshaler interface.
func (t *Table) MarshalJSON() ([]byte, error) {
	players := map[string]*PlayerState{}
	for seat, player := range t.Players() {
		players[strconv.FormatInt(int64(seat), 10)] = player
	}

	tJSON := &tableJSON{
		Options:     t.opts,
		Deck:        t.deck,
		Button:      t.Button(),
		Action:      t.Action(),
		Round:       t.Round(),
		MinRaise:    t.MinRaise(),
		Board:       t.Board(),
		Players:     players,
		Pot:         t.Pot(),
		StartedHand: t.startedHand,
	}
	return json.Marshal(tJSON)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Table) UnmarshalJSON(b []byte) error {
	tJSON := &tableJSON{}
	if err := json.Unmarshal(b, tJSON); err != nil {
		return err
	}

	players := map[int]*PlayerState{}
	for seat, player := range tJSON.Players {
		i, err := strconv.ParseInt(seat, 10, 64)
		if err != nil {
			return err
		}
		players[int(i)] = player
	}

	seats := &Seats{
		players:  players,
		minBuyin: tJSON.Options.MinBuyin,
		maxBuyin: tJSON.Options.MaxBuyin,
	}

	t.opts = tJSON.Options
	t.dealer = hand.NewDealer()
	t.deck = tJSON.Deck
	t.button = tJSON.Button
	t.action = tJSON.Action
	t.round = tJSON.Round
	t.minRaise = tJSON.MinRaise
	t.board = tJSON.Board
	t.seats = seats
	t.pot = tJSON.Pot
	t.startedHand = tJSON.StartedHand

	return nil
}

func (t *Table) setUpHand() {
	t.deck = t.dealer.Deck()
	t.round = 0
	t.button = t.seats.next(t.button+1, false)
	t.action = -1
	t.pot = pot.New(t.NumOfSeats())

	// reset cards
	t.board = []*hand.Card{}
	for _, player := range t.seats.Players() {
		player.holeCards = []*HoleCard{}
		player.out = false
		player.allin = false
	}
}

func (t *Table) setUpRound() {
	// deal board cards
	bCards := t.game().BoardCards(t.deck, round(t.round))
	t.board = append(t.board, bCards...)
	t.resetActed()

	for seat, player := range t.seats.Players() {
		// add hole cards
		hCards := t.game().HoleCards(t.deck, round(t.round))
		player.holeCards = append(player.holeCards, hCards...)

		// add forced bets
		pos := t.seats.relativePos(t.button, seat)
		chips := t.game().ForcedBet(t.holeCards(), t.opts, round(t.round), seat, pos)
		t.addToPot(seat, chips)
	}

	// set starting action position
	relativePos := t.game().RoundStartSeat(t.holeCards(), round(t.round))
	seat := (relativePos + t.button) % t.NumOfSeats()
	t.action = t.seats.next(seat, true)

	// set raise amounts
	t.minRaise = t.opts.Stakes.BigBet
	t.resetCanRaise(-1)

	// if everyone is all in, skip round
	count := 0
	for _, player := range t.players {
		if !player.allin && !player.out {
			count++
		}
	}
	if count < 2 {
		t.action = -1
	}
}

func (t *Table) payoutResults(resultsMap map[int][]*pot.Result) {
	for seat, results := range resultsMap {
		for _, result := range results {
			amount := t.players[seat].chips + result.Chips
			p := t.players[seat]
			p.chips = amount
			t.players[seat] = p
		}
	}
}

func (t *Table) handleAction(seat int, p *PlayerState, a Action, chips int) error {
	// validate action
	validAction := false
	for _, va := range t.ValidActions() {
		validAction = validAction || va == a
	}
	if !validAction {
		return ErrInvalidAction
	}

	// check if bet or raise amount is invalid
	if (a == Bet || a == Raise) && (chips < t.MinRaise() || chips > t.MaxRaise()) {
		switch a {
		case Bet:
			return ErrInvalidBet
		case Raise:
			return ErrInvalidRaise
		}
	}

	// commit action
	switch a {
	case Fold:
		p.out = true
	case Call:
		t.addToPot(seat, t.Outstanding())
	case Bet:
		t.addToPot(seat, chips)
		t.resetActed()
		if chips >= t.minRaise {
			t.resetCanRaise(seat)
			t.minRaise = chips
		}
	case Raise:
		t.addToPot(seat, chips+t.Outstanding())
		t.resetActed()
		if chips >= t.minRaise {
			t.resetCanRaise(seat)
			t.minRaise = chips
		}
	}
	p.canRaise = false
	p.acted = true
	return nil
}

func (t *Table) addToPot(seat, chips int) {
	p := t.players[seat]
	if chips >= p.chips {
		chips = p.chips
		p.allin = true
	}
	p.chips -= chips
	t.pot.Contribute(seat, chips)
}

func (t *Table) hasNextHand() bool {
	count := 0
	for _, player := range t.players {
		if player.chips > 0 {
			count++
		}
	}
	return count > 1
}

func (t *Table) holeCards() map[int][]*HoleCard {
	hCards := map[int][]*HoleCard{}
	for seat, player := range t.players {
		hCards[seat] = player.holeCards
	}
	return hCards
}

func (t *Table) resetActed() {
	for _, player := range t.players {
		player.acted = false
	}
}

func (t *Table) resetCanRaise(seat int) {
	for s, player := range t.players {
		player.canRaise = !(s == seat)
	}
}

func (t *Table) everyoneFolded() bool {
	count := 0
	for _, player := range t.players {
		if !player.out {
			count++
		}
	}
	return count < 2
}

func (t *Table) game() game {
	return t.opts.Game.get()
}

func isNil(o interface{}) bool {
	return o == nil || !reflect.ValueOf(o).Elem().IsValid()
}
