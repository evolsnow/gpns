package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	pb "github.com/evolsnow/gpns/protos"
	"github.com/evolsnow/samaritan/common/log"
	"github.com/gorilla/websocket"
	apns "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"golang.org/x/net/context"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"sync"
	"time"
)

// server is used to implement rpc.GPNSServer.
type server struct{}

// SayHello implements rpc.GPNSServer
func (s server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name, Age: 24}, nil
}

// SocketPush push msg to user specified tokens with websocket
func (s server) SocketPush(ctx context.Context, in *pb.SocketPushRequest) (*pb.SocketPushReply, error) {
	log.Info("calling socket push")
	var offline []string
	payload := make(map[string]string)
	payload["message"] = in.Message
	for k, v := range in.ExtraInfo {
		payload[k] = v
	}
	raw, _ := json.Marshal(payload)
	for _, ut := range in.UserToken {
		var (
			v  *websocket.Conn
			ok bool
		)
		if v, ok = socketConnMap[ut]; !ok {
			offline = append(offline, ut)
			continue
		}
		err := v.WriteMessage(websocket.TextMessage, raw)
		if err != nil {
			//client closed
			offline = append(offline, ut)
		}
	}

	return &pb.SocketPushReply{UserToken: offline}, nil
}

// ApplePush push msg with apns
func (s server) ApplePush(ctx context.Context, in *pb.ApplePushRequest) (*pb.ApplePushReply, error) {
	log.Info("calling apple push")
	cert, err := certificate.FromPemFile("dev.pem", "")
	if err != nil {
		log.Error(err)
	}
	client := apns.NewClient(cert).Development()

	payload := payload.NewPayload()
	payload.Alert(in.Message)
	payload.Sound("default")
	payload.Badge(1)
	for k, v := range in.ExtraInfo {
		payload.Custom(k, v)
	}
	reply := new(pb.ApplePushReply)
	var wg sync.WaitGroup
	wg.Add(len(in.DeviceToken))
	for _, token := range in.DeviceToken {
		nf := new(apns.Notification)
		nf.DeviceToken = token

		nf.Payload = payload
		go func(*apns.Notification) {
			defer wg.Done()
			resp, _ := client.Push(nf)
			if resp.StatusCode != 200 {
				log.Error("push notification error:", resp.Reason)
				reply.DeviceToken = append(reply.DeviceToken, nf.DeviceToken)
			} else {
				log.Info("successfully push:", nf.DeviceToken)
			}
		}(nf)
	}
	wg.Wait()
	return reply, nil
}

// SendMail send code with gmail
func (s server) SendMail(ctx context.Context, in *pb.MailRequest) (*pb.MailResponse, error) {
	log.Info("calling send mail")
	now := time.Now()
	//smtpServer := "mail.samaritan.tech"
	//auth := smtp.PlainAuth(
	//	"",
	//	"user",
	//	"password",
	//	smtpServer,
	//)
	smtpServer := "smtp.gmail.com"
	auth := smtp.PlainAuth(
		"",
		"godo.noreply@gmail.com",
		"password",
		smtpServer,
	)
	from := mail.Address{Name: "GoDo", Address: "godo.noreply@gmail.com"}
	to := mail.Address{Name: in.To, Address: in.To}
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
		smtpServer+":587",
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	)

	return new(pb.MailResponse), err
}

// SendSMS send sms to mobile with yunpian
func (s server) SendSMS(ctx context.Context, in *pb.SMSRequest) (*pb.SMSResponse, error) {
	log.Info("calling send sms")
	apiKey := "apikey"
	ypURL := "https://sms.yunpian.com/v1/sms/send.json"
	resp, err := http.PostForm(ypURL, url.Values{"apikey": {apiKey}, "mobile": {in.To}, "text": {in.Text}})
	if err != nil {
		log.Error(err.Error())
	}
	defer resp.Body.Close()
	type ypReply struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	reply := new(ypReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	if reply.Code != 0 {
		return &pb.SMSResponse{Success: false, Reason: reply.Msg}, nil
	}
	return &pb.SMSResponse{Success: true}, nil
}

// ReceiveMsg receive Chat Msg from app
func (s server) ReceiveMsg(in *pb.ReceiveChatRequest, stream pb.GPNS_ReceiveMsgServer) error {
	for {
		c := <-chats
		msg := &pb.ReceiveChatReply{Chat: c}
		stream.Send(msg)
	}
}
