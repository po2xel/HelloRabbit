package hub

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"github.com/satori/go.uuid"
)


type RPCClient struct {
	conn *amqp.Connection
	channel *amqp.Channel
	queue amqp.Queue
}


type message struct {
	id   uuid.UUID
	body []byte
}


// Hub maintains the set of active clients
type Hub struct {
	// Connected clients
	clients map[uuid.UUID]*Client

	// Inbound message from the clients
	recv chan *message

	// Outbound message to the clients
	sent chan *message

	// Connect requests from the clients
	conn chan *Client

	// Disconnect request from the clients
	disconn chan *Client

	rpc *RPCClient
}


func newRpcServer() *RPCClient {
	conn, err :=  amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Printf("error: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("error: %v", err)
	}

	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Printf("error: %v", err)
	}

	return &RPCClient{
		conn: conn,
		channel: ch,
		queue: q,
	}
}

func New() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]*Client),
		recv:    make(chan *message),
		sent: 	 make(chan *message),
		conn:    make(chan *Client),
		disconn: make(chan *Client),
		rpc:     newRpcServer(),
	}
}


func (rpc *RPCClient) consume(sent chan *message) {
	msgs, err := rpc.channel.Consume(
		rpc.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Printf("error: %v", err)
	}

	for msg := range msgs {
		log.Println("Received recv RPC message")

		if id, err := uuid.FromString(msg.CorrelationId); err != nil {
			log.Printf("UUID conversion error: %v", err)
		} else {
			sent <- &message{id: id, body: msg.Body}
		}
	}

}

func (rpc *RPCClient) publish(msg *message) {
	if err := rpc.channel.Publish("", "rpc_queue", false, false,
		amqp.Publishing{ContentType: "image/png", Body: msg.body, ReplyTo: rpc.queue.Name, CorrelationId: msg.id.String()});
		err != nil {
		log.Printf("error: %v", err)
	} else {
		log.Println("Message published")
	}
}


func (hub *Hub) Run() {
	go hub.rpc.consume(hub.sent)

	for {
		select {
		case client := <-hub.conn:
			log.Println("Connected")
			hub.clients[client.id] = client

		case client := <-hub.disconn:
			if _, ok := hub.clients[client.id]; ok {
				delete(hub.clients, client.id)
				close(client.send)

				log.Println("Disconnected")
			}

		case msg := <-hub.recv:
			// log.Printf("Received messge length:  %v", len(recv))
			hub.rpc.publish(msg)

		case msg := <-hub.sent:
			hub.clients[msg.id].send <- msg.body
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}


// ServeWS handles websocket request from the front end.
func (hub *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, maxSendQueue), id: uuid.NewV4()}
	client.hub.conn <- client

	go client.WriteMessage()
	go client.ReadMessage()

	//hub.rpc.channel.Close()
	//hub.rpc.conn.Close()
}
