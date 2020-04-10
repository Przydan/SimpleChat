package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	// forward - channel containing sended msg
	// which should send to user web browser
	forward chan []byte

	// the join is a channel for clients who wants join to room
	join chan *client

	// channel for those who wants leave room
	leave chan *client

	// clients contains all clients who are currently in this room
	clients map[*client]bool
}

// method newRoom create new room ready to use
func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// join client
			r.clients[client] = true
		case client := <-r.leave:
			// leave room
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			// sending msg for all clients
			for client := range r.clients {
				client.send <- msg
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
