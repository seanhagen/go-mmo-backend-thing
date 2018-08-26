package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func sendStateToWS(gs <-chan gameState, cn *websocket.Conn, cnLock *sync.Mutex, erCh chan<- error, player *player) {
	log.Print("waiting for state to send to websocket")
	for {
		select {
		case state := <-gs:
			log.Printf("##### need to send state to websocket")
			// create the state message to send
			outMsg := stateMessage{
				Type:  "state",
				Board: state.board,
			}

			timeout := time.Now().Add(5 * time.Second)
			cn.SetWriteDeadline(timeout)

			// try to send state to the user
			log.Printf("##### locking websocket to send")
			cnLock.Lock()
			err := cn.WriteJSON(outMsg)

			log.Printf("##### message sent")
			cnLock.Unlock()
			log.Printf("##### websocket unlocked")

			// if there's an error writing to the websocket, close the websocket and  send
			// an error on the error channel erCh
			if err != nil {
				erCh <- err
				errStr := fmt.Sprintf("##### error sending game sate message: %v", err)
				errMsg := wsMessage{Msg: errStr}
				_ = cn.WriteJSON(errMsg)
				log.Print(errStr)
				break
			}

		default:
		}
	}
}

func handleWSIncoming(pm chan<- playerMsg, cn *websocket.Conn, cnLock *sync.Mutex, erCh chan<- error, player *player) {
	log.Printf("handling incoming for player %v", player.id)
	for {
		// try to read in the incoming message
		log.Printf("$$$$ handle incoming websocket: read message")
		msgType, data, err := cn.ReadMessage()

		// if there's an error reading from the websocket, close the websocket and
		// send an error to the error channel erCh
		log.Printf("$$$$ handle incoming websocket: got message -- type %v, data: %v", msgType, data)
		if err != nil {
			errStr := fmt.Sprintf("$$$$ handle incoming websocket: error reading message: %v", err)
			log.Print(errStr)

			erCh <- err
			errMsg := wsMessage{Msg: errStr}
			_ = cn.WriteJSON(errMsg)
			break
		}

		// try to unmarshal what should be json
		msg := &wsMessage{}
		err = json.Unmarshal(data, msg)
		if err != nil {
			errStr := fmt.Sprintf("unable to parse json: %v", err)
			log.Printf("unable to parse json: %v", err)

			erCh <- err
			errMsg := wsMessage{Msg: errStr}
			_ = cn.WriteJSON(errMsg)
			break
		}

		pm <- playerMsg{
			player: *player,
			action: msg.Msg,
		}
	}

	// there was an error reading from the websocket, so we're done
	log.Printf("incoming done for player %v", player.id)
}

func sendMessagesToWS(ms <-chan string, cn *websocket.Conn, cnLock *sync.Mutex, erCh chan<- error) {
	log.Print("waiting for messages to send back to the player")
	for {
		select {
		case msg := <-ms:
			outMsg := wsMessage{
				Type: "response",
				Msg:  msg,
			}
			// try to send the message to the player
			cnLock.Lock()
			err := cn.WriteJSON(outMsg)
			cnLock.Unlock()

			// if there's an error writing to the websocket, close the websocket and  send
			// an error on the error channel erCh
			if err != nil {
				erCh <- err
				errStr := fmt.Sprintf("error sending update message: %v", err)
				errMsg := wsMessage{Msg: errStr}
				_ = cn.WriteJSON(errMsg)
				log.Print(errStr)
				break
			}
		}
	}
}
