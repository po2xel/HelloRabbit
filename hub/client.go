package hub

import (
	"github.com/gorilla/websocket"
	"time"
	"log"
	"github.com/satori/go.uuid"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	maxSendQueue = 256
)

// Client is a middleman between the websocket connection and the hub
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound message
	send chan []byte

	// Client Id
	id uuid.UUID
}


// ReadMessage pumps messages from the websocket connection to the hub
func (client *Client) ReadMessage() {
	defer func() {
		client.hub.disconn <- client
		client.conn.Close()
	}()

	// client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(
		func(string) error {
			client.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
	})

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			log.Printf("error: %v", err)
			break
		}

		//dst, err := base64.StdEncoding.DecodeString(string(recv[22:]))
		//log.Printf("error: %v", err)
		//err = ioutil.WriteFile("test.png", dst, 0666)
		//log.Printf("error: %v", err)
		// recv = bytes.TrimSpace(bytes.Replace(recv, newline, space, -1))
		//err = ioutil.WriteFile("test.png", recv, os.ModeType)
		//log.Printf("error: %v", err)

		//client.hub.recv <- dst
		client.hub.recv <- &message{client.id, msg}
	}
}


// WriteMessage pumps messages from the hub to the websocket connection
func (client *Client) WriteMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// The hub closed the channel
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				client.conn.Close()
				break
			} else {
				log.Println("Message sent")
			}

		case <- ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}