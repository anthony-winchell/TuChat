package main

import (
  "net"
	"bufio"
	"os"
	"log"
)

func main() {
	//connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Println(err)
		return 
	}
	defer conn.Close()

	//get connection message and username prompt
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return 
	}
	log.Println(string(buffer[:n]))

	//get username from the user 
	terminalReader := bufio.NewReader(os.Stdin)
	username, err := terminalReader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return 
	}

	//send username to the server
	_, err = conn.Write([]byte(username))
	if err != nil {
		log.Println(err)
		return 
	}

	//read messages from terminal and send to server 
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