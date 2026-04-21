
# FluxKV ⚡️

A high-performance, concurrent, and RESP-compatible distributed key-value store built entirely from scratch in Go.

FluxKV is designed to be a lightweight, Redis-like in-memory database. It natively implements the Redis Serialization Protocol (RESP), ensuring seamless compatibility with standard Redis clients (like `redis-cli`) while leveraging Go's powerful concurrency model to handle tens of thousands of simultaneous connections.

## Key Features

- **RESP Compatible:** Implements a custom parser for the Redis Serialization Protocol. Standard Redis clients and benchmarking tools work out of the box.
- **Highly Concurrent:** Utilizes Go's lightweight goroutines to handle multiple TCP client connections without event-loop blocking.
- **Thread-Safe Core:** Protects shared memory access using efficient read/write mutexes (`sync.RWMutex`) to guarantee data integrity under heavy load.
- **AOF Persistence:** Ensures data durability by logging raw mutating commands to an Append-Only File. Data survives server restarts.
- **Distributed Replication:** Supports Leader-Follower architecture. Features full state transfer on connection and real-time broadcasting of write commands to all active replicas.
- **Zero Dependencies:** Built entirely using the Go standard library (`net`, `bufio`, `sync`, `os`).

---

## Architecture & Project Structure

FluxKV is strictly modularized to separate networking, parsing, storage, and cluster management.

```text
fluxkv/
├── cmd/server/main.go          # TCP listener, CLI flags, and connection router
├── internal/
│   ├── aof/aof.go              # File I/O for Append-Only File persistence
│   ├── replication/repl.go     # Leader-Follower state transfer & broadcasting
│   ├── resp/resp.go            # Custom byte parser/encoder for RESP
│   └── store/store.go          # Thread-safe concurrent memory map
```

---

## Getting Started

### Prerequisites
- **Go 1.20+**: Required to compile the server.
- **Redis CLI**: Required to interact with the database (`redis-cli` and `redis-benchmark`).

### 1. Clone & Build
```bash
git clone [https://github.com/Fashad-Ahmed/fluxkv.git](https://github.com/Fashad-Ahmed/fluxkv.git)
cd fluxkv

# Compile the Go application
go build -o fluxkv-server ./cmd/server
```

### 2. Run a Standalone Node
By default, the server listens on port `6379`.
```bash
./fluxkv-server
```

---

## Testing & Usage Guide

Once the server is running, open a new terminal window and use `redis-cli` to interact with it.

### Basic Interaction
```text
$ redis-cli -p 6379
127.0.0.1:6379> PING
PONG
127.0.0.1:6379> SET ai_model "Proximal Policy Optimization"
OK
127.0.0.1:6379> GET ai_model
"Proximal Policy Optimization"
```

### Testing Durability (Crash Recovery)
FluxKV saves your data to `database_6379.aof`. 
1. Write some data using the `SET` command.
2. Kill the FluxKV server using `Ctrl+C`.
3. Start the server again (`./fluxkv-server`). It will print `Reading database_6379.aof to restore state...`
4. Reconnect with `redis-cli` and `GET` your key. Your data will still be there.

### Testing Distributed Replication
Test the Master-Replica syncing by running two instances. You will need three terminal windows.

**Terminal 1 (Start the Leader):**
```bash
./fluxkv-server --port=6379
```

**Terminal 2 (Start the Follower):**
```bash
./fluxkv-server --port=6380 --replicaof=localhost:6379
```

**Terminal 3 (The Client):**
Write to the Leader and verify the data replicates to the Follower.
```bash
# Write data to the Leader
$ redis-cli -p 6379 SET architecture "distributed"
OK

# Read that exact data from the Follower
$ redis-cli -p 6380 GET architecture
"distributed"
```

---

## Performance Benchmarking

FluxKV is built for speed. You can test its limits using the official `redis-benchmark` tool. Ensure a single node is running on port 6379, then run:

```bash
# Send 100,000 parallel requests to the server
redis-benchmark -p 6379 -t set,get -n 100000 -q
```

*Expected Output:*
```text
SET: 85470.09 requests per second, p50=0.211 msec
GET: 88105.73 requests per second, p50=0.203 msec
```

---

## License

This project is licensed under the MIT License.
