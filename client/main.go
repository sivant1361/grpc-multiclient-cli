package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	message "go-cli-chat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const port = ":5050"

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	sentence, _ := reader.ReadString('\n')
	sentence = strings.TrimSpace(sentence)
	return sentence
}

func ChatClientFunc(client message.MessageServiceClient) {
	fmt.Println("\nClient Cli")
	fmt.Println("**********")
	fmt.Println()

	fmt.Print("Enter username: ")

	username := getInput()
	connect := username

	stream, err := client.SendReply(context.Background())
	if err != nil {
		log.Fatalf("could not send reply: %v", err)
	}

	go func() {
		for {
			message, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while streaming %v", err)
			}

			from := message.GetFrom()
			to := message.GetTo()
			content := message.GetContent()

			if from == to {
				fmt.Print("\r", "server> ", content, "\n", username+"> ")
				connect = username
			} else {
				fmt.Print("\r", from+"> ", content, "\n", username+"> ")
			}
		}
	}()

	fmt.Print(username + "> ")
	sendAll := false
	for {
		inp := getInput()

		if strings.Contains(inp, "connect") {
			sendAll = false
			if strings.Contains(inp, "disconnect") {
				connect = username
			} else {
				connect = strings.Split(inp, " ")[1]
				fmt.Print("message> ")
				inp = getInput()
			}
		} else if strings.Contains(inp, "sendall") {
			sendAll = true
			fmt.Print("message> ")
			inp = getInput()
		} else if inp == "exit" {
			fmt.Println("")
			break
		}
		req := &message.Message{
			Content:   inp,
			From:      username,
			To:        connect,
			Broadcast: sendAll,
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("Error while sending %v", err)
		}
		fmt.Print(username + "> ")

		time.Sleep(time.Second / 2)
	}
	stream.CloseSend()
	log.Printf("Chating with server is finished")
}

func main() {
	conn, err := grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	c := message.NewMessageServiceClient(conn)

	log.Println("Initial connect done")
	ChatClientFunc(c)

}
