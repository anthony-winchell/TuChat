package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"tuchat/protocol"
	"strings"
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

	displayWelcome(decoder)

	if err := setUsername(decoder, encoder, terminalReader); err != nil {
		log.Println(err)
		return
	}

	done := make(chan struct{})
	go func() {
		receiveMessages(reader)
		close(done)
	}()

	go sendMessages(conn, terminalReader)

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
	}
}

// handles sending messages to the server
func sendMessages(conn net.Conn, terminalReader *bufio.Reader) {
	for {
		message, err := terminalReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		_, err = conn.Write([]byte(message))
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func displayWelcome(decoder *json.Decoder) {
	var message protocol.Message

	if err := decoder.Decode(&message); err != nil {
		log.Println(err)
		return
	}
	fmt.Println(message.Message)
}

// after connecting to the server, the client will be prompted to enter a username
func setUsername(decoder *json.Decoder, encoder *json.Encoder, terminalReader *bufio.Reader) error {
	var message protocol.Message
	for {
		if err := decoder.Decode(&message); err != nil {
			return err
		}

		switch message.Type {
			case "username_prompt":
				fmt.Println(message.Message)

				username, err := terminalReader.ReadString('\n')
				if err != nil {
					return err
				}

				if err := Send(encoder, protocol.Message{
					Type: "username",
					Username: strings.TrimSpace(username),
				}); err != nil {
					return err	
				}
			case "error":
				fmt.Println(message.Message)
			case "username_accepted": 
				return nil
		}
	}
}

func Send(encoder *json.Encoder, message protocol.Message) error {
	return encoder.Encode(message)
}
