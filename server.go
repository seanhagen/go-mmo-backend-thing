package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func newServer() *gameServer {
	board := make([]row, boardSize)
	for i := 0; i < boardSize; i++ {
		board[i] = make([]tile, boardSize)
	}

	state := &gameState{
		board: board,
	}

	return &gameServer{
		notifier: make(chan gameState, 1),

		state: state,
		sl:    &sync.Mutex{},

		playerMessages: make(chan playerMsg),

		nextPlayerID: 1,
		npidLock:     &sync.Mutex{},

		players: map[chan gameState]bool{},

		closingClients: make(chan chan gameState),
		newClients:     make(chan chan gameState),

		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (gs *gameServer) listen() {
	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		// a client connected, do the thing
		case newClient := <-gs.newClients:
			log.Printf("player connected")
			gs.players[newClient] = true

		case clientLeft := <-gs.closingClients:
			log.Printf("player disconnected")
			delete(gs.players, clientLeft)

		case stateUpdate := <-gs.notifier:
			log.Printf("### need to update clients of game state update")
			for c := range gs.players {
				log.Printf("### sending state to channel: %#v", c)
				c <- stateUpdate
				log.Printf("### sent state to channel: %#v", c)
			}
			log.Printf("### clients notified of game state update")

		case playerUpdate := <-gs.playerMessages:
			log.Printf("got message from player")
			// when we get a player message
			go func(update playerMsg, state *gameState, stateLock *sync.Mutex) {
				log.Printf("hey, player wants to do a thing")
				stateLock.Lock()
				log.Printf("update from player: %v", update)
				stateLock.Unlock()
			}(playerUpdate, gs.state, gs.sl)

		case tick := <-ticker.C:
			// every tick, notify all clients about the state
			log.Printf("server tick: %v", tick)
			gs.sl.Lock()
			gs.notifier <- *gs.state
			gs.sl.Unlock()

		default:
		}
	}

}
