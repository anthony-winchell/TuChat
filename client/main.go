package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	//connect to the server
	conn, err := connectToServer()
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	reader := bufio.NewReader(conn)
	terminalReader := bufio.NewReader(os.Stdin)

	if err := setUsername(conn, reader, terminalReader); err != nil {
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
func receiveMessages(reader *bufio.Reader) {
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(strings.TrimSpace(message))
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

// after connecting to the server, the client will be prompted to enter a username
func setUsername(conn net.Conn, reader *bufio.Reader, terminalReader *bufio.Reader) error {

	for {
		prompt, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		prompt = strings.TrimSpace(prompt)
		fmt.Println(prompt)

		if strings.HasPrefix(prompt, "Thank you") {
			return nil
		}

		if strings.HasSuffix(prompt, ":") {
			username, err := terminalReader.ReadString('\n')
			if err != nil {
				return err
			}

			_, err = conn.Write([]byte(username))
			if err != nil {
				return err
			}
		}
	}
}
