package main

import (
	"flag"
	"fmt"
	"github.com/evolsnow/httprouter"
	pb "github.com/evolsnow/gpns/protos"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

var port = flag.Int("p", 10086, "The rpc server port")

func main() {
	flag.Parse()
	// http server
	router := httprouter.New()
	router.GET("/websocket/:token", Socket)
	log.Fatal(http.ListenAndServe(":10000", router))

	// rpc server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGPNSServer(s, &server{})
	log.Println("listen on", fmt.Sprintf(":%d", *port))
	s.Serve(lis)
}
