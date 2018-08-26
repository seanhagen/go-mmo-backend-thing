package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

const boardSize = 10

// tile level
// how much gold on this level
// percentage chance of finding gold
type tileLevel struct {
	gold       int
	findChance float32
}

// tile
type tile struct {
	// lock mutex
	lock *sync.Mutex

	// what player is on this tile (only one player per tile?)
	currentPlayer int

	// what level that player is on. starts at 0 for ground level,
	// going down to -10 for the lowest level
	playerLevel int

	// tile levels
	levels []tileLevel
}

type row []tile

// state -- updated by server when it receives player messages
type gameState struct {
	// the game board
	board []row
}

// player -- someone playing in the game
type player struct {
	// locking mutex
	lock *sync.Mutex

	// player id. duh
	id int

	// how much gold this player has
	gold int

	// how much of a chance this player has to find gold. when digging down,
	// the find chance is (level find chance + player find chance), a number less
	// than 0.95 and greater than 0. when a player finds gold, this is increased
	// by (level/1000) -- finding gold on lower levels increases the chances more
	// than finding gold on the higher level
	findChance float32

	// channel to send success/error message to (gets sent back to user via websocket)
	msgs chan string
}

// player message -- created when message recieved from websocket
type playerMsg struct {
	// player
	player player

	// player action (move,dig,climb,etc)
	action string
}

// messages sent to or received from the player
type wsMessage struct {
	Type string `json:"type"`
	// the message the player sent
	Msg string `json:"message"`
}

// game state messages sent to the player
type stateMessage struct {
	Type  string `json:"type"`
	Board []row
}

type errorMessage struct {
	Error string `json:"error"`
}

// server -- the thing doing all the work
type gameServer struct {
	// when the game state needs to be sent out, it's sent to this channel.
	// when this channel gets a message, all clients in players will get the updated
	// game state
	notifier chan gameState

	// the current state of the game world. it's updated whenever a player sends a
	// valid message that can change the state of the game world
	state *gameState

	// mutex to lock the game state
	sl *sync.Mutex

	// channel that recieves updates from players and updates the state of the game world
	playerMessages chan playerMsg

	// the id of the next player to join
	nextPlayerID int
	// mutex to lock reading/writing the next player id
	npidLock *sync.Mutex

	// all the connected players, how they get sent the current game state
	players map[chan gameState]bool

	// when a client disconnects, this channel gets a message
	closingClients chan chan gameState

	// when a client connects, this channel gets a message
	newClients chan chan gameState

	// upgrader, turns http request into websocket
	upgrader websocket.Upgrader
}
