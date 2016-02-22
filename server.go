package main

import (
	"github.com/anachronistic/apns"
	pb "github.com/evolsnow/gpns/protos"
	"golang.org/x/net/context"
	"log"
	"sync"
)

// server is used to implement rpc.GreeterServer.
type server struct{}

// SayHello implements rpc.GreeterServer
func (s server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name, Age: 24}, nil
}

// ApplePush
func (s server) ApplePush(ctx context.Context, in *pb.PushRequest) (*pb.PushReply, error) {
	payload := apns.NewPayload()
	payload.Alert = in.Message
	payload.Sound = "default"
	payload.Badge = 1
	client := apns.NewClient("gateway.sandbox.push.apple.com:2195", "cert.pem", "key.pem")

	reply := new(pb.PushReply)
	var wg sync.WaitGroup
	wg.Add(len(in.DeviceToken))
	for _, token := range in.DeviceToken {
		pn := apns.NewPushNotification()
		pn.DeviceToken = token
		for k, v := range in.ExtraInfo {
			pn.Set(k, v)
		}
		pn.AddPayload(payload)
		go func(*apns.PushNotification) {
			defer wg.Done()
			resp := client.Send(pn)
			if resp.Error != nil {
				log.Println("push notification error:", resp.Error)
				reply.DeviceToken = append(reply.DeviceToken, pn.DeviceToken)
			} else {
				log.Println("successfully push:", pn.DeviceToken)
				reply.Count++
			}
		}(pn)
	}
	wg.Wait()
	return reply, nil
}

//// Stream Apple Push
//func (s server) StreamPush(stream pb.Greeter_StreamPushServer)  error {
//	reply := new(pb.PushReply)
//	payload := apns.NewPayload()
//	payload.Sound = "default"
//	payload.Badge = 1
//	client := apns.NewClient("gateway.sandbox.push.apple.com:2195", "cert.pem", "key.pem")
//	for {
//		pushInfo, err := stream.Recv()
//		if err == io.EOF {
//			return stream.SendAndClose(reply)
//		}
//		if err != nil {
//			return err
//		}
//		payload.Alert = pushInfo.Message
//		pn := apns.NewPushNotification()
//		pn.DeviceToken = pushInfo.DeviceToken
//		for k, v := range pushInfo.ExtraInfo {
//			pn.Set(k, v)
//		}
//		pn.AddPayload(payload)
//		go func(*apns.PushNotification) {
//			resp := client.Send(pn)
//			if resp.Error != nil {
//				log.Println("push notification error:", resp.Error)
//				reply.DeviceToken = append(reply.DeviceToken, pn.DeviceToken)
//			} else {
//				log.Println("successfully push:", pn.DeviceToken)
//				reply.Count++
//			}
//		}(pn)
//	}
//}
