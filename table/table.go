package table

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/SyntropyDev/joker/hand"
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

// Player represents a player at a table.
type Player interface {
	// ID returns the unique identifier of the player.
	ID() string

	// FromID resets the player from an id.  It is required for
	// deserialization.
	FromID(id string) (Player, error)

	// Action returns the action and it's chip amount.  This method
	// will block table's Next() function until input is recieved.
	Action() (a Action, chips int)
}

// RegisterPlayer stores the player implementation for json deserialization.
func RegisterPlayer(p Player) {
	registeredPlayer = p
}

// RegisterDeck stores the deck implementation for json deserialization.
func RegisterDeck(d hand.Deck) {
	registeredDeck = d
}

var (
	// mapping to player implemenation
	registeredPlayer Player

	// mapping to deck implementation
	registeredDeck hand.Deck
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
	FixedLimit = "FL"
)

// NumOfSeats is a configuration for how many seats are
// available at the table.
type NumOfSeats int

const (
	// TwoSeats forms a 2-handed table
	TwoSeats NumOfSeats = iota + 2

	// ThreeSeats forms a 3-handed table
	ThreeSeats

	// FourSeats forms a 4-handed table
	FourSeats

	// FiveSeats forms a 5-handed table
	FiveSeats

	// SixSeats forms a 6-handed table
	SixSeats

	// SevenSeats forms a 7-handed table
	SevenSeats

	// EightSeats forms a 8-handed table
	EightSeats

	// NineSeats forms a 9-handed table
	NineSeats

	// TenSeats forms a 10-handed table
	TenSeats
)

// Options are the configurations for creating a table.
type Options struct {

	// Game is the game of the table.
	Game Game `json:"game"`

	// Limit is the limit of the table
	Limit Limit `json:"limit"`

	// Stakes is the stakes for the table.
	Stakes Stakes `json:"stakes"`

	// NumOfSeats is the number of seats available for the table.
	NumOfSeats NumOfSeats `json:"numbOfSeats"`
}

// PlayerState is the state of a player at a table.
type PlayerState struct {
	player    Player
	holeCards []*HoleCard
	chips     int
	acted     bool
	out       bool
	allin     bool
	canRaise  bool
}

// Acted returns whether or not the player has acted for the current round.
func (state *PlayerState) Acted() bool {
	return state.acted
}

// AllIn returns whether or not the player is all in for the current hand.
func (state *PlayerState) AllIn() bool {
	return state.allin
}

// CanRaise returns whether or not the player can raise in the current round.
func (state *PlayerState) CanRaise() bool {
	return state.canRaise
}

// Chips returns the number of chips the player has in his or her stack.
func (state *PlayerState) Chips() int {
	return state.chips
}

// HoleCards returns the hole cards the player currently has.
func (state *PlayerState) HoleCards() []*HoleCard {
	c := []*HoleCard{}
	return append(c, state.holeCards...)
}

// Out returns whether or not the player is out of the current hand.
func (state *PlayerState) Out() bool {
	return state.out
}

// Player returns the player.
func (state *PlayerState) Player() Player {
	return state.player
}

// String returns a string useful for debugging.
func (state *PlayerState) String() string {
	const format = "{Player: %s, HoleCards: %v, Chips: %d, Acted: %t, Out: %t, AllIn: %t}"
	return fmt.Sprintf(format,
		state.player.ID(), state.holeCards, state.chips, state.acted, state.out, state.allin)
}

type playerStateJSON struct {
	ID        string      `json:"id"`
	HoleCards []*HoleCard `json:"holeCards"`
	Chips     int         `json:"chips"`
	Acted     bool        `json:"acted"`
	Out       bool        `json:"out"`
	Allin     bool        `json:"allin"`
	CanRaise  bool        `json:"canRaise"`
}

// MarshalJSON implements the json.Marshaler interface.
func (state *PlayerState) MarshalJSON() ([]byte, error) {
	tpJSON := &playerStateJSON{
		ID:        state.Player().ID(),
		HoleCards: state.HoleCards(),
		Chips:     state.Chips(),
		Acted:     state.Acted(),
		Out:       state.Out(),
		Allin:     state.AllIn(),
		CanRaise:  state.CanRaise(),
	}
	return json.Marshal(tpJSON)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (state *PlayerState) UnmarshalJSON(b []byte) error {
	tpJSON := &playerStateJSON{}
	if err := json.Unmarshal(b, tpJSON); err != nil {
		return err
	}

	if isNil(registeredPlayer) {
		return errors.New("table: PlayerState json deserialization requires use of the RegisterPlayer function")
	}

	p, err := registeredPlayer.FromID(tpJSON.ID)
	if err != nil {
		return fmt.Errorf("table PlayerState json deserialization failed because of player %s FromID - %s", tpJSON.ID, err)
	}

	state.player = p
	state.holeCards = tpJSON.HoleCards
	state.chips = tpJSON.Chips
	state.acted = tpJSON.Acted
	state.out = tpJSON.Out
	state.allin = tpJSON.Allin
	state.canRaise = tpJSON.CanRaise

	return nil
}

// Table represent a poker table and dealer.  A table manages the
// game state and all player interactions at the table.
type Table struct {
	opts        Options
	deck        hand.Deck
	button      int
	action      int
	round       int
	minRaise    int
	board       []*hand.Card
	players     map[int]*PlayerState
	pot         *Pot
	startedHand bool
}

// New creates a new table with the options and deck provided.  To
// start playing hands, at least two players must be seated and the
// Next() function must be called.  If the number of seats is invalid
// for the Game specified New panics.
func New(opts Options, deck hand.Deck) *Table {
	if int(opts.NumOfSeats) > opts.Game.get().MaxSeats() {
		format := "table: %s has a maximum of %d seats but attempted %d"
		s := fmt.Sprintf(format, opts.Game, opts.Game.get().MaxSeats(), opts.NumOfSeats)
		panic(s)
	}

	return &Table{
		opts:    opts,
		deck:    deck,
		board:   []*hand.Card{},
		players: map[int]*PlayerState{},
		pot:     newPot(int(opts.NumOfSeats)),
		action:  -1,
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
	return t.players[t.Action()]
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
// -1 is returned.
func (t *Table) MaxRaise() int {
	player := t.CurrentPlayer()
	if isNil(player) {
		return -1
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
// -1 is returned.
func (t *Table) MinRaise() int {
	player := t.CurrentPlayer()
	if isNil(player) {
		return -1
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
// current player.  If there is no current player then -1 is returned.
func (t *Table) Outstanding() int {
	player := t.CurrentPlayer()
	if isNil(player) {
		return -1
	}
	if player.AllIn() || player.Out() {
		return 0
	}
	return t.pot.outstanding(t.Action())
}

// Players returns a mapping of seats to player states.  Empty seats
// are not included.
func (t *Table) Players() map[int]*PlayerState {
	players := map[int]*PlayerState{}
	for seat, p := range t.players {
		players[seat] = p
	}
	return players
}

// View returns a view of the table that only contains information
// privileged to the given player.
func (t *Table) View(p Player) *Table {
	players := map[int]*PlayerState{}
	for seat, player := range t.players {
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

	return &Table{
		opts:        t.opts,
		deck:        hand.EmptyDeck(),
		button:      t.button,
		action:      t.action,
		round:       t.round,
		minRaise:    t.minRaise,
		board:       t.board,
		pot:         t.pot,
		startedHand: t.startedHand,
		players:     players,
	}
}

// Pot returns the current pot.
func (t *Table) Pot() *Pot {
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
func (t *Table) Next() (results map[int][]*PotResult, done bool, err error) {
	if !t.startedHand {
		if !t.hasNextHand() {
			return nil, true, NewInsufficientPlayers()
		}
		t.setUpHand()
		t.setUpRound()
		t.startedHand = true
		return nil, false, nil
	}

	if t.action == -1 {
		t.round++

		if t.round == t.game().NumOfRounds() {
			highHands := newHands(t.holeCards(), t.board, t.game().FormHighHand)
			lowHands := newHands(t.holeCards(), t.board, t.game().FormLowHand)
			results = t.pot.payout(highHands, lowHands, t.game().Sorting(), t.game().SplitPot(), t.button)
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
		for seat, player := range t.players {
			if player.out {
				continue
			}
			results = t.pot.take(seat)
			t.payoutResults(results)
			t.startedHand = false
			return results, false, nil
		}
	}

	t.action = t.nextSeat(t.action+1, true)
	return nil, false, nil
}

// Sit sits the player at the table with the given amount of chips.
// An error is return if the seat is invalid, the player is already
// seated, the seat is already occupied, or the chips are outside
// the valid buy in amounts.
func (t *Table) Sit(p Player, seat, chips int) error {
	if !t.validSeat(seat) {
		return NewInvalidSeat(seat)
	} else if t.isSeated(p) {
		return NewAlreadySeated(p.ID())
	} else if _, occupied := t.players[seat]; occupied {
		return NewSeatOccupied(seat)
	}

	min := (t.opts.Stakes.SmallBet * 50)
	max := (t.opts.Stakes.SmallBet * 200)
	if chips < min || chips > max {
		return NewInvalidBuyIn(chips)
	}

	t.players[seat] = &PlayerState{
		player:    p,
		holeCards: []*HoleCard{},
		chips:     chips,
	}
	return nil
}

// Stand removes the player from the table.  If the player isn't
// seated the command is ignored.
func (t *Table) Stand(p Player) {
	for seat, pl := range t.players {
		if pl.player.ID() == p.ID() {
			delete(t.players, seat)
			return
		}
	}
}

type tableJSON struct {
	Options     Options                 `json:"options"`
	Cards       []*hand.Card            `json:"cards"`
	Discards    []*hand.Card            `json:"discards"`
	Button      int                     `json:"button"`
	Action      int                     `json:"action"`
	Round       int                     `json:"round"`
	MinRaise    int                     `json:"minRaise"`
	Board       []*hand.Card            `json:"board"`
	Players     map[string]*PlayerState `json:"players"`
	Pot         *Pot                    `json:"pot"`
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
		Cards:       t.deck.Cards(),
		Discards:    t.deck.Discards(),
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
	if isNil(registeredDeck) {
		return errors.New("table: Deck json deserialization requires use of the RegisterDeck function")
	}

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

	t.opts = tJSON.Options
	t.deck = registeredDeck.FromCards(tJSON.Cards, tJSON.Discards)
	t.button = tJSON.Button
	t.action = tJSON.Action
	t.round = tJSON.Round
	t.minRaise = tJSON.MinRaise
	t.board = tJSON.Board
	t.players = players
	t.pot = tJSON.Pot
	t.startedHand = tJSON.StartedHand

	return nil
}

func (t *Table) setUpHand() {
	t.deck.Reset()
	t.round = 0
	t.button = t.nextSeat(t.button+1, false)
	t.action = -1
	t.pot = newPot(t.NumOfSeats())

	// reset cards
	t.board = []*hand.Card{}
	for _, player := range t.players {
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

	for seat, player := range t.players {
		// add hole cards
		hCards := t.game().HoleCards(t.deck, round(t.round))
		player.holeCards = append(player.holeCards, hCards...)

		// add forced bets
		pos := t.relativePosition(seat)
		chips := t.game().ForcedBet(t.holeCards(), t.opts, round(t.round), seat, pos)
		t.addToPot(seat, chips)
	}

	// set starting action position
	relativePos := t.game().RoundStartSeat(t.holeCards(), round(t.round))
	seat := (relativePos + t.button) % t.NumOfSeats()
	t.action = t.nextSeat(seat, true)

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

func (t *Table) payoutResults(resultsMap map[int][]*PotResult) {
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
		return NewInvalidAction(a, t.ValidActions())
	}

	// check if bet or raise amount is invalid
	if (a == Bet || a == Raise) && (chips < t.MinRaise() || chips > t.MaxRaise()) {
		switch a {
		case Bet:
			return NewInvalidBet(chips, t.MinRaise(), t.MaxRaise())
		case Raise:
			return NewInvalidRaise(chips, t.MinRaise(), t.MaxRaise())
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
	t.pot.contribute(seat, chips)
}

func (t *Table) nextSeat(seat int, playing bool) int {
	count := 0
	seat = seat % t.NumOfSeats()
	for count < t.NumOfSeats() {
		p, ok := t.players[seat]
		if ok && (!playing || (!p.out && !p.allin && !p.acted)) {
			return seat
		}
		count++
		seat = (seat + 1) % t.NumOfSeats()
	}
	return -1
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

func (t *Table) isSeated(p Player) bool {
	for _, pl := range t.players {
		if p.ID() == pl.player.ID() {
			return true
		}
	}
	return false
}

func (t *Table) validSeat(seat int) bool {
	return seat >= 0 && seat < t.NumOfSeats()
}

func (t *Table) relativePosition(seat int) int {
	current := t.button
	count := 0
	for {
		if current == seat {
			break
		}
		current = t.nextSeat(current+1, false)
		count++
	}
	return count
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
