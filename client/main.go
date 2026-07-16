package main

import (
  "net"
	"bufio"
	"log"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Println(err)
		return 
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println(err)
		return 
	}
	log.Println(string(buffer[:n]))


	_, err = conn.Write([]byte("Hello from client"))
	if err != nil {
		log.Println(err)
		return 
	}
}