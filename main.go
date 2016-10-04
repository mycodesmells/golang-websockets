package main

import (
	"fmt"
	"net/http"
	"strings"
)

type Message struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

func main() {
	http.HandleFunc("/broadcast/", broadcastHandler)
	http.Handle("/ws", wsHandler)

	http.ListenAndServe(":3000", nil)
}

func broadcastHandler(w http.ResponseWriter, r *http.Request) {
	msg := readMsgFromRequest(r)
	broadcast(&Message{"Server", msg})
	fmt.Fprintf(w, "Broadcasting %v", msg)
}

func readMsgFromRequest(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	return parts[2]
}
