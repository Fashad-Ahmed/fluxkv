package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Fashad-Ahmed/fluxkv/internal/aof"
	"github.com/Fashad-Ahmed/fluxkv/internal/resp"
	"github.com/Fashad-Ahmed/fluxkv/internal/store"
)

func main() {
	kv := store.NewMemoryStore()

	// Initialize AOF
	aof, err := aof.NewAof("database.aof")
	if err != nil {
		log.Fatalf("Failed to open AOF file: %v", err)
	}
	defer aof.Close()

	// Replay AOF on startup
	fmt.Println("Reading AOF file to restore state...")
	aof.Read(func(value resp.Value) {
		command := strings.ToUpper(value.Array[0].Str)
		args := value.Array[1:]

		if command == "SET" && len(args) == 2 {
			kv.Set(args[0].Str, args[1].Str)
		}
	})

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

		// Pass BOTH the store and the AOF to the handler
		go handleConnection(conn, kv, aof)
	}
}

func handleConnection(conn net.Conn, kv *store.MemoryStore, aof *aof.Aof) {
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

		if value.Typ != "array" || len(value.Array) == 0 {
			continue
		}

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
			
			// Save to memory
			kv.Set(args[0].Str, args[1].Str)
			
			// Append the exact command to the AOF log
			aof.Write(value)
			
			conn.Write([]byte("+OK\r\n"))
			
		case "GET":
			if len(args) != 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue
			}
			val, exists := kv.Get(args[0].Str)
			if !exists {
				conn.Write([]byte("$-1\r\n"))
			} else {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
			}
			
		default:
			conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", command)))
		}
	}
}