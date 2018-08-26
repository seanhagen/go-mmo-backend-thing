package main

import (
	"log"
	"net/http"
	"sync"
)

// the handler for websockets
func (gs *gameServer) handle(w http.ResponseWriter, r *http.Request) {
	// when a websocket connection request is made
	// upgrade the connection to an actual websocket
	conn, err := gs.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error upgrading to websocket: %v", err)
		return
	}

	// create a lock for reading from or writing from websocket
	wsLock := &sync.Mutex{}

	// get the current player id, then increment the next player id
	gs.npidLock.Lock()
	pid := gs.nextPlayerID
	log.Printf("adding player: %v", pid)
	gs.nextPlayerID++
	gs.npidLock.Unlock()

	// create channel for getting messages that need to be sent back to the player
	pch := make(chan string)

	// create a new player
	p := &player{
		lock:       &sync.Mutex{},
		gold:       0,
		findChance: 0.2,
		msgs:       pch,
	}

	// create channel gsUpdates to receive game state updates
	// create lock for reading from or writing to the websocket
	gsUpdates := make(chan gameState)

	// send gsUpdates channel to gs.newClients
	gs.newClients <- gsUpdates

	// create channel that gets an error if the websocket is closed for whatever
	// reason. basically, if we can't read from or write to the websocket, we're done
	wsErr := make(chan error)

	log.Printf("all ready to launch gofuncs")

	// launch a goroutine to read in messages from the websocket
	// and then do the first stage of validation. if validation passes
	// then send message to gs.playerMessages channel
	go handleWSIncoming(gs.playerMessages, conn, wsLock, wsErr, p)
	log.Printf("incoming weboscket message handler launched")

	// launch goroutine to wait for messages that need to be sent back to the player
	go sendMessagesToWS(pch, conn, wsLock, wsErr)
	log.Printf("send message to websocket handler launched")

	// launch goroutine to wait for game state messages from gsUpdates
	// and send it out on the websocket
	go sendStateToWS(gsUpdates, conn, wsLock, wsErr, p)
	log.Printf("send state to websocket handler launched")

	// if the websocket is disconnected, send a message to gs.closingClients
	err = <-wsErr
	log.Printf("websocket error: %v, closing handler now", err)
	gs.closingClients <- gsUpdates
}
