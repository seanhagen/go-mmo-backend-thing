package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

const boardSize = 50

type tile struct {
}

type gameState struct {
	nextPlayerId int
	npidmu       *sync.Mutex

	board [boardSize][boardSize]tile
}

type playerInMsg struct {
	Type string
}

type playerOutMsg struct {
	Msg string
}

type gameStateMsg struct {
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	playerIds := 1
	pIDmu := &sync.Mutex{}

	// gs := gameState{}
	// gsChan := make(chan gameState)
	// gsMu := &sync.Mutex{}

	// outState := map[int]chan gameState{}
	outMsgChans := map[int]chan playerOutMsg{}
	inMsgChans := map[int]chan *playerInMsg{}

	indexFile, err := os.Open("templates/index.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	index, err := ioutil.ReadAll(indexFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = indexFile.Close()

	jsFile, err := os.Open("templates/index.js")
	if err != nil {
		fmt.Println(err)
		return
	}
	js, err := ioutil.ReadAll(jsFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = jsFile.Close()

	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		pIDmu.Lock()
		defer pIDmu.Unlock()

		oc := make(chan playerOutMsg)
		ic := make(chan *playerInMsg)

		outMsgChans[playerIds] = oc
		inMsgChans[playerIds] = ic

		playerIds = playerIds + 1

		go setupWS(conn, ic, oc)
	})

	http.HandleFunc("/index.js", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(js))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(index))
	})

	http.ListenAndServe(":3000", nil)
}

func handleGameState(in chan playerInMsg, out chan playerOutMsg, gs gameState) {

}

func setupWS(conn *websocket.Conn, in chan *playerInMsg, out chan playerOutMsg) {
	fmt.Println("Client subscribed")

	shutdown := make(chan int)

	conn.SetCloseHandler(func(code int, t string) error {
		close(shutdown)
		fmt.Printf("hey thing closed: %v %v\n", code, t)
		return nil
	})

	go handlePlayerMsg(conn, in)
	go sendPlayerMsg(conn, out, shutdown)
}

func sendPlayerMsg(conn *websocket.Conn, out chan playerOutMsg, shutdown chan int) {
	for {
		select {
		case msg, more := <-out:
			conn.WriteJSON(msg)
			if !more {
				return
			}
		case <-shutdown:
			return
		}
	}
}

func handlePlayerMsg(conn *websocket.Conn, in chan *playerInMsg) {
	for {
		msg := &playerInMsg{}
		err := conn.ReadJSON(msg)

		if err != nil {
			fmt.Printf("handleWS, error: %v\n", err)
			break
		}

		in <- msg
	}
	// conn.Close()
	// fmt.Println("Client unsubscribed")
}
