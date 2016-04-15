package main

import (
	"encoding/json"
	"flag"
	"fmt"
	pb "github.com/evolsnow/gpns/protos"
	"github.com/evolsnow/httprouter"
	"github.com/evolsnow/samaritan/common/log"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

var cfg *Config

func main() {
	var conf string
	var err error
	flag.StringVar(&conf, "c", "config.json", "specify config file")
	flag.Parse()
	cfg, err = parseConfig(conf)
	if err != nil {
		log.Fatal("a vailid json config file must exist")
	}
	// http server
	router := httprouter.New()
	router.GET("/websocket", webSocket)
	go func() {
		log.Fatal(http.ListenAndServe(":10000", router))
	}()

	// rpc server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal("failed to listen", err)
	}
	s := grpc.NewServer()
	pb.RegisterGPNSServer(s, &server{})
	log.Info("listen on", fmt.Sprintf(":%d", cfg.Port))
	s.Serve(lis)
}

// rpc config
type Config struct {
	Port         int    `json:"port,omitempty"`
	Cert         string `json:"cert"`
	CertPassword string `json:"cert_password,omitempty"`
	Production   bool   `json:"production,omitempty"`
	MailPassword string `json:"mail_password,omitempty"`
	ApiKey       string `json:"api_key,omitempty"`
}

// ParseConfig parses config from the given file path
func parseConfig(path string) (config *Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	config = &Config{}
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return
}
