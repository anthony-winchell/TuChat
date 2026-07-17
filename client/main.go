package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

func main() {
	//connect to the server
	conn, err := connectToServer()
	if err != nil {
		log.Println(err)
		return
	}

	terminalReader := bufio.NewReader(os.Stdin)

	if err := setUsername(conn, terminalReader); err != nil {
		log.Println(err)
		return 
	}
	
	go receiveMessages(conn)

	sendMessages(conn, terminalReader)

}

func connectToServer() (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", "localhost:8080")
	return 
}

//handles recieving messages from the server. use as goroutine so that it can run in the background
func receiveMessages(conn net.Conn) {
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(buffer[:n]))
	}
}

//handles sending messages to the server 
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

//after connecting to the server, the client will be prompted to enter a username
func setUsername(conn net.Conn, terminalReader *bufio.Reader) error {
	//get username prompt
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return err
	}
	log.Println(string(buffer[:n]))

	//get username from terminal input
	username, err := terminalReader.ReadString('\n')
	if err != nil {
		return err
	}

	//send username to the server
	_, err = conn.Write([]byte(username))
	if err != nil {
		return err
	}

	return nil
}
