package main

import (
	"log"
	"net"
	"io"
)

type Client struct {
	Username string
	Conn     net.Conn
}

func main() {
	listener, err := net.Listen("tcp",":8080")
	if err != nil {
		log.Println(err)
		return 
	}
	defer listener.Close()
	log.Println("Listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
		defer conn.Close()

		_, err := conn.Write([]byte("Connection Successful\n Enter Username: "))
		if err != nil {
			log.Println(err)
			return
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}
		username := string(buffer[:n])

		for {
			n, err := conn.Read(buffer)
			if err != nil {
				if err == io.EOF {
					log.Println("Client Disconnected")
					return 
				}
				log.Println(err)
				return
			}

			message := string(buffer[:n])
			log.Printf("%s: %s", username, message)
		}
	}