package main

import (
	"fmt"
	"go-im/websocket"
	"net/http"
)

func main() {
	http.HandleFunc("/ws", ws.WSHandler)
	http.HandleFunc("/push", ws.PushHandler)
	http.HandleFunc("/broad", ws.BroadcastHandler)
	if err := http.ListenAndServe(":8888", nil); err != nil {
		fmt.Println(err.Error())
	}
}
