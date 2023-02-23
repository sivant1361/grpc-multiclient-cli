package main

import (
	message "go-cli-chat/proto"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
)

type messageServer struct {
	message.MessageServiceServer
}

var channelMapper = make(map[string]chan *message.Message)
var ActiveMapper = make(map[string]bool)

// var messageQueue []string

func SenderChat(stream message.MessageService_SendReplyServer, ch chan *message.Message) error {
	for {
		value, ok := <-ch
		if ok {
			from := value.GetFrom()
			to := value.GetTo()
			content := value.GetContent()

			res := &message.Message{
				Content: content,
				From:    from,
				To:      to,
			}

			if err := stream.Send(res); err != nil {
				return err
			}
		}
	}
}

func Broadcast(stream message.MessageService_SendReplyServer, req *message.Message, username string) {
	for user := range channelMapper {
		if user == username {
			continue
		}
		req.To = user
		channelMapper[user] <- req
	}
}

func (s *messageServer) SendReply(stream message.MessageService_SendReplyServer) error {

	flag := false

	for {
		req, err := stream.Recv()
		if flag == false {
			log.Printf("Got connected with : %v", req.GetFrom())
			channelMapper[req.GetFrom()] = make(chan *message.Message)
			go SenderChat(stream, channelMapper[req.GetFrom()])
			ActiveMapper[req.GetFrom()] = true
			flag = true
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("Request from : %v", req.GetFrom())
		if req.GetBroadcast() {
			Broadcast(stream, req, req.GetFrom())
			continue
		}

		if ActiveMapper[req.GetTo()] != true {
			req.Content = "Sender named " + req.To + " not there"
			req.To = req.GetFrom()
		}

		channelMapper[req.GetTo()] <- req
		// messageQueue = append(messageQueue, req.Content)
	}
}

const port = ":5050"

func main() {
	lis, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	gs := grpc.NewServer()

	message.RegisterMessageServiceServer(gs, &messageServer{})
	log.Printf("Server started at %v", lis.Addr())

	if err := gs.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}
