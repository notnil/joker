//go:build js && wasm

// Command wasm is the browser client for a 9-max cash table vs bots.
//
//	make -C cmd/wasm
//	go run ./cmd/server
//
// Then open http://localhost:8080
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/notnil/joker/pkg/holdem"
	"github.com/notnil/joker/pkg/play"
)

var game *play.Game

type dto struct {
	holdem.View
	Log   []string `json:"log"`
	Error string   `json:"error,omitempty"`
}

func main() {
	js.Global().Set("jokerNewGame", js.FuncOf(jsNewGame))
	js.Global().Set("jokerState", js.FuncOf(jsState))
	js.Global().Set("jokerAct", js.FuncOf(jsAct))
	js.Global().Set("jokerStep", js.FuncOf(jsStep))
	js.Global().Set("jokerReady", js.ValueOf(true))
	select {}
}

func jsNewGame(this js.Value, args []js.Value) any {
	g, err := play.NewCash(0, nil)
	if err != nil {
		return marshal(dto{Error: err.Error()})
	}
	game = g
	return marshal(snapshot())
}

func jsState(this js.Value, args []js.Value) any {
	if game == nil {
		return marshal(dto{Error: "no game"})
	}
	return marshal(snapshot())
}

func jsAct(this js.Value, args []js.Value) any {
	if game == nil {
		return marshal(dto{Error: "no game"})
	}
	typ := ""
	chips := 0
	if len(args) > 0 {
		typ = args[0].String()
	}
	if len(args) > 1 && !args[1].IsUndefined() && !args[1].IsNull() {
		chips = args[1].Int()
	}
	a, err := holdem.ParseAction(typ, chips)
	if err != nil {
		return marshal(dto{Error: err.Error()})
	}
	if err := game.Act(a); err != nil {
		d := snapshot()
		d.Error = err.Error()
		return marshal(d)
	}
	return marshal(snapshot())
}

func jsStep(this js.Value, args []js.Value) any {
	if game == nil {
		return marshal(dto{Error: "no game"})
	}
	if err := game.Step(); err != nil {
		d := snapshot()
		d.Error = err.Error()
		return marshal(d)
	}
	return marshal(snapshot())
}

func snapshot() dto {
	return dto{View: game.View(), Log: game.Log()}
}

func marshal(d dto) string {
	b, err := json.Marshal(d)
	if err != nil {
		return `{"error":"marshal"}`
	}
	return string(b)
}
