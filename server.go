package main

import (
	"encoding/base64"
	"fmt"
	"github.com/anachronistic/apns"
	pb "github.com/evolsnow/gpns/protos"
	"github.com/evolsnow/samaritan/common/log"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"net/mail"
	"net/smtp"
	"sync"
	"time"
)

// server is used to implement rpc.GPNSServer.
type server struct{}

// SayHello implements rpc.GPNSServer
func (s server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name, Age: 24}, nil
}

// Socket push
func (s server) SocketPush(ctx context.Context, in *pb.SocketPushRequest) (*pb.SocketPushReply, error) {
	offline := make([]string, len(in.UserToken))
	for _, ut := range in.UserToken {
		var (
			v  *websocket.Conn
			ok bool
		)
		if v, ok = socketConnMap[ut]; !ok {
			offline = append(offline, ut)
			continue
		}
		err := v.WriteMessage(websocket.TextMessage, []byte(in.Message))
		if err != nil {
			//client closed
			offline = append(offline, ut)
		}
	}

	return &pb.SocketPushReply{UserToken: offline}, nil
}

// ApplePush
func (s server) ApplePush(ctx context.Context, in *pb.ApplePushRequest) (*pb.ApplePushReply, error) {
	payload := apns.NewPayload()
	payload.Alert = in.Message
	payload.Sound = "default"
	payload.Badge = 1
	client := apns.NewClient("gateway.sandbox.push.apple.com:2195", "cert.pem", "key.pem")

	reply := new(pb.ApplePushReply)
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
			}
		}(pn)
	}
	wg.Wait()
	return reply, nil
}

// Mail send
func (s server) SendMail(ctx context.Context, in *pb.MailRequest) (*pb.MailResponse, error) {
	now := time.Now()
	smtpServer := "mail.samaritan.tech"
	auth := smtp.PlainAuth(
		"",
		"admin",
		"passwdforadmin",
		smtpServer,
	)

	from := mail.Address{"Samaritan", "admin@mail.samaritan.tech"}
	to := mail.Address{"收件人", in.To}
	title := in.Subject
	body := in.Body

	header := make(map[string]string)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	header["List-Unsubscribe"] = ""

	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = encodeRFC2047(title)
	header["Date"] = now.Format("Mon, _2 Jan 2006 15:04:05 +0800 (CST)") //"Mon, 1 Mar 2016 10:51:00 +0800 (CST)"
	header["Message-Id"] = makeMessageId("mail.samaritan.tech")

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		smtpServer+":25",
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	)

	return new(pb.MailResponse), err
}

// Chat Msg from app
func (s server) ReceiveMsg(in *pb.ReceiveChatRequest, stream pb.GPNS_ReceiveMsgServer) error {
	for {
		c := <-chats
		msg := &pb.ReceiveChatReply{Chat: c}
		stream.Send(msg)
	}
}
