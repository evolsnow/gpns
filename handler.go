package main

import (
	"github.com/evolsnow/httprouter"
	"github.com/evolsnow/samaritan/common/log"
	"github.com/gorilla/websocket"
	"github.com/nu7hatch/gouuid"
	"net/http"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	return true
}} // use default options for webSocket

var socketConnMap = make(map[string]*websocket.Conn)
var chats = make(chan string, 100)

//keep deviceToken and connection
func webSocket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ah := r.Header.Get("Authorization")
	if ah != "" {
		//iOS
		//if dt := ps.ByName("deviceToken"); dt != "" {
		//	deviceMap[ah] = dt
		//}
	} else {
		u4, _ := uuid.NewV4()
		ah = u4.String()
	}
	establishSocketConn(w, r, ah)
}

func establishSocketConn(w http.ResponseWriter, r *http.Request, ut string) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn("upgrade:", err)
		return
	}
	socketConnMap[ut] = c
	log.Info("new socket conn:", ut)
	defer c.Close()
	defer delete(socketConnMap, ut)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Warn("read:", err)
			break
		}
		log.Debug("rec: %s", message)
		go handlerMsg(message)
	}
}

func handlerMsg(msg []byte) {
	chats <- string(msg)
}
