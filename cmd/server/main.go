package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Fashad-Ahmed/fluxkv/internal/resp"
	"github.com/Fashad-Ahmed/fluxkv/internal/store"
)

func main() {
	// Initialize our thread-safe memory store
	kv := store.NewMemoryStore()

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatalf("Failed to bind to port 6379: %v", err)
	}
	defer listener.Close()

	fmt.Println("FluxKV is running on port 6379...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go handleConnection(conn, kv)
	}
}

func handleConnection(conn net.Conn, kv *store.MemoryStore) {
	defer conn.Close()
	r := resp.NewResp(conn)

	for {
		value, err := r.Read()
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Printf("Connection error: %s\n", err.Error())
			}
			return
		}

		// Redis commands are sent as arrays of bulk strings
		if value.Typ != "array" || len(value.Array) == 0 {
			continue
		}

		// Extract the command (e.g., "SET", "GET") and its arguments
		command := strings.ToUpper(value.Array[0].Str)
		args := value.Array[1:]

		switch command {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
			
		case "SET":
			if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				continue
			}
			// Save to our memory map
			kv.Set(args[0].Str, args[1].Str)
			conn.Write([]byte("+OK\r\n"))
			
		case "GET":
			if len(args) != 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue
			}
			
			// Retrieve from our memory map
			val, exists := kv.Get(args[0].Str)
			if !exists {
				// RESP standard for "Null" or "Not Found"
				conn.Write([]byte("$-1\r\n"))
			} else {
				// RESP standard for a returned Bulk String
				response := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
				conn.Write([]byte(response))
			}
			
		default:
			conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", command)))
		}
	}
}