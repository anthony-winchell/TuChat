package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"tuchat/protocol"
)

func main() {
	//connect to the server
	conn, err := connectToServer()
	if err != nil {
		log.Println(err)
		return
	}
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	defer conn.Close()

	terminalReader := bufio.NewReader(os.Stdin)

	if err := authenticate(decoder, encoder, terminalReader); err != nil {
		log.Println(err)
		return
	}

	done := make(chan struct{})
	go func() {
		receiveMessages(decoder)
		close(done)
	}()

	go sendMessages(encoder, terminalReader)

	<-done

	fmt.Println("Disconnected from server")
}

func connectToServer() (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", "localhost:8080")
	return
}

// handles recieving messages from the server. use as goroutine so that it can run in the background
func receiveMessages(decoder *json.Decoder) {
	for {
		var message protocol.Message

		if err := decoder.Decode(&message); err != nil {
			log.Println(err)
			return
		}

		switch message.Type {
		case "chat":
			fmt.Printf("%s: %s\n", message.Username, message.Message)
		case "pm":
			fmt.Printf("[PM] %s: %s\n", message.Username, message.Message)
		case "system":
			fmt.Println(message.Message)
		case "users":
			renderUsers(message.Users)
		case "welcome":
			fmt.Println(message.Message)
		case "auth_prompt":
			fmt.Println(message.Message)
		case "join":
			fmt.Printf("%s joined the chat\n", message.Username)
		case "leave":
			fmt.Printf("%s left the chat\n", message.Username)
		case "error":
			fmt.Println("Error: ", message.Message)
		}

	}
}

// handles sending messages to the server
func sendMessages(encoder *json.Encoder, terminalReader *bufio.Reader) {
	for {
		input, err := terminalReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		msg := protocol.Message{
			Message: input,
		}

		if strings.HasPrefix(input, "/") {
			msg.Type = "command"
		} else {
			msg.Type = "chat"
		}

		if err := Send(encoder, msg); err != nil {
			log.Println(err)
			return
		}
	}
}

// after connecting to the server, the client will be prompted to enter a username
func authenticate(decoder *json.Decoder, encoder *json.Encoder, terminalReader *bufio.Reader) error {
	for {
		var message protocol.Message

		if err := decoder.Decode(&message); err != nil {
			return err
		}

		switch message.Type {

		case "auth_prompt": 
			if err := promptAndSendAuth(encoder, terminalReader); err != nil {
				return err
			}
		case "error":
			fmt.Println("Error: ", message.Message)
			if err := promptAndSendAuth(encoder, terminalReader); err != nil {
				return err
			}

		case "auth_success": 
			return nil
		}
	}
}

func promptAndSendAuth(encoder *json.Encoder, terminalReader *bufio.Reader) error {
	for {
		fmt.Println("1. Login")
		fmt.Println("2. Register")

		fmt.Print("> ")

		choice, err := terminalReader.ReadString('\n')
		if err != nil {
			return err 
		}

		choice = strings.TrimSpace(choice)

		msgType := ""

		switch choice {
		case "1": 
			msgType = "login"

		case "2":
			msgType = "register"

		default:
			fmt.Println("Invalid option")
			continue
		}

		fmt.Print("Username: ")

		username, err := terminalReader.ReadString('\n')
		if err != nil {
			return err
		}

		fmt.Print("Password: ")

		password, err := terminalReader.ReadString('\n')
		if err != nil {
			return err
		}

		return Send(encoder, protocol.Message{
			Type:     msgType,
			Username: strings.TrimSpace(username),
			Password: strings.TrimSpace(password),
		})
	}
}

func renderUsers(usernames []string) {
	fmt.Println("Users in this Room: ")
	for _, username := range usernames {
		fmt.Println("*" + username)
	}
}

func Send(encoder *json.Encoder, message protocol.Message) error {
	return encoder.Encode(message)
}
