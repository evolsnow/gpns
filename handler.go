package main

import (
	"github.com/evolsnow/httprouter"
	"github.com/gorilla/websocket"
	"github.com/nu7hatch/gouuid"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	return true
}} // use default options for webSocket

var socketMap = make(map[string]*websocket.Conn)

func establishWebSocketConn(w http.ResponseWriter, r *http.Request, uid string) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	socketMap[uid] = c
	log.Println("new socket conn:", uid)
	defer c.Close()
	defer delete(socketMap, uid)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
	}
}

func RawSocket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	u4, _ := uuid.NewV4()
	uid := u4.String()
	establishWebSocketConn(w, r, uid)
}

func AuthedSocket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := ps.Get("token")
	establishWebSocketConn(w, r, uid)
}
