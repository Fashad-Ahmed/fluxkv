package main

import (
	"flag" 
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Fashad-Ahmed/fluxkv/internal/aof"
	"github.com/Fashad-Ahmed/fluxkv/internal/resp"
	"github.com/Fashad-Ahmed/fluxkv/internal/store"
	"github.com/Fashad-Ahmed/fluxkv/internal/replication"
)

func main() {
	port := flag.String("port", "6379", "Port to run the server on")
	replicaOf := flag.String("replicaof", "", "Connect to a leader (e.g., localhost:6379)")
	flag.Parse()

	kv := store.NewMemoryStore()

	// Use the dynamic port for the AOF file so the Leader and Follower 
	// don't try to write to the exact same file!
	aofFile := fmt.Sprintf("database_%s.aof", *port)
	aofStore, err := aof.NewAof(aofFile)
	if err != nil {
		log.Fatalf("Failed to open AOF file: %v", err)
	}
	defer aofStore.Close()

	fmt.Printf("Reading %s to restore state...\n", aofFile)
	aofStore.Read(func(value resp.Value) {
		command := strings.ToUpper(value.Array[0].Str)
		args := value.Array[1:]
		if command == "SET" && len(args) == 2 {
			kv.Set(args[0].Str, args[1].Str)
		}
	})

	if *replicaOf != "" {
		fmt.Printf("Starting as Follower. Replicating from: %s\n", *replicaOf)
		go connectToLeader(*replicaOf, kv, aofStore)
	}

	address := fmt.Sprintf(":%s", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to bind to port %s: %v", *port, err)
	}
	defer listener.Close()

	fmt.Printf("FluxKV is running on port %s...\n", *port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go handleConnection(conn, kv, aofStore)
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

func connectToLeader(leaderAddr string, kv *store.MemoryStore, aofStore *aof.Aof) {
	conn, err := net.Dial("tcp", leaderAddr)
	if err != nil {
		log.Fatalf("Failed to connect to leader at %s: %v", leaderAddr, err)
	}
	
	fmt.Println("Successfully connected to Leader.")

	// Let the leader know we are a replica by sending a special command
	conn.Write([]byte("*1\r\n$4\r\nSYNC\r\n"))

	// Listen for commands forwarded by the Leader forever
	r := resp.NewResp(conn)
	for {
		value, err := r.Read()
		if err != nil {
			log.Printf("Lost connection to leader: %v\n", err)
			return
		}

		if value.Typ != "array" || len(value.Array) == 0 {
			continue
		}

		command := strings.ToUpper(value.Array[0].Str)
		args := value.Array[1:]

		// A Follower should only ever receive write commands from the Leader
		if command == "SET" && len(args) == 2 {
			// Update local memory
			kv.Set(args[0].Str, args[1].Str)
			// Append to the Follower's AOF log
			aofStore.Write(value)
			fmt.Printf("Replicated SET command from Leader: %s = %s\n", args[0].Str, args[1].Str)
		}
	}
}