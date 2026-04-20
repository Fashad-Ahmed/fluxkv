package main

import (
	"fmt"
	"net"

	"github.com/Fashad-Ahmed/fluxkv/internal/resp"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("New client connected: %s\n", conn.RemoteAddr().String())

	// Create our new RESP reader passing in the TCP connection
	r := resp.NewResp(conn)

	for {
		value, err := r.Read()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Client disconnected.")
			} else {
				fmt.Printf("Error reading from client: %s\n", err.Error())
			}
			return
		}

		fmt.Printf("Parsed Command: %+v\n", value.Array)

		conn.Write([]byte("+OK\r\n"))
	}
}

func main() {
	fmt.Println("Starting FluxKV server on :6379")

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Printf("Failed to bind to port 6379: %s\n", err.Error())
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %s\n", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}