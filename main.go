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

var clients []Client

func main() {
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})

	http.HandleFunc("/broadcast/", func(w http.ResponseWriter, r *http.Request) {
		msg := readMsgFromRequest(r)
		broadcast(&Message{"Server", msg})
		fmt.Fprintf(w, "Broadcasting %v", msg)
	})

	http.Handle("/ws", wsHandler)

	http.ListenAndServe(":3000", nil)
}

func readMsgFromRequest(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	return parts[2]
}
