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

func Socket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	uid := ps.Get("token")
	if uid == "" {
		u4, _ := uuid.NewV4()
		uid = u4.String()
	}
	socketMap[uid] = c
	defer c.Close()
	defer delete(socketMap, uid)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		for i := 0; i < 10; i++ {
			err = c.WriteMessage(mt, []byte("fuck"))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}
