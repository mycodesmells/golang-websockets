package main

import (
	"fmt"

	"golang.org/x/net/websocket"
)

var wsHandler = websocket.Handler(onWsConnect)

func onWsConnect(ws *websocket.Conn) {
	defer ws.Close()
	client := NewClient(ws)
	clients = addClientAndGreet(clients, client)
	client.listen()
}

func broadcast(msg *Message) {
	fmt.Printf("Broadcasting %+v\n", msg)
	for _, c := range clients {
		c.ch <- msg
	}
}

func addClientAndGreet(list []Client, client Client) []Client {
	clients = append(list, client)
	websocket.JSON.Send(client.connection, Message{"Server", "Welcome!"})
	return clients
}
