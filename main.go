package main

import (
	"flag"
	"fmt"
	"github.com/evolsnow/httprouter"
	"github.com/evolsnow/samaritan/common/log"
	pb "github.com/evolsnow/samaritan/gpns/protos"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

var port = flag.Int("p", 10086, "The rpc server port")

func main() {
	flag.Parse()
	// http server
	router := httprouter.New()
	router.GET("/websocket", webSocket)
	go func() {
		log.Fatal(http.ListenAndServe(":10000", router))
	}()

	// rpc server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal("failed to listen", err)
	}
	s := grpc.NewServer()
	pb.RegisterGPNSServer(s, &server{})
	log.Println("listen on", fmt.Sprintf(":%d", *port))
	s.Serve(lis)
}
