# 🏗️ FluxKV Architecture & Design Document

This document outlines the internal architecture, data flow, and design decisions behind **FluxKV**, a distributed, concurrent key-value store built in Go.

---

## 1. System Overview

FluxKV is an in-memory database that adheres to the Redis Serialization Protocol (RESP). It is designed to handle high-throughput concurrent connections while guaranteeing data durability and offering high availability through Leader-Follower replication.

### Core Guiding Principles:
1. **Concurrency over Event Loops:** Instead of using a single-threaded asynchronous event loop (like Redis/Node.js), FluxKV leverages Go's lightweight goroutines to handle multiplexing, allowing multiple CPU cores to process requests simultaneously.
2. **Standard Protocol Compatibility:** By natively parsing RESP, FluxKV integrates transparently with existing Redis ecosystem tools (`redis-cli`, `redis-benchmark`).
3. **Simplicity in Persistence:** Data durability is achieved through a human-readable Append-Only File (AOF) log, avoiding complex binary formats.

---

## 2. Component Architecture

The system is strictly modularized into distinct domains to ensure separation of concerns.

### `cmd/server` (The Entry Point & Router)
- **Role:** Handles TCP socket binding, parses command-line configuration (`--port`, `--replicaof`), and manages the lifecycle of client connections.
- **Mechanism:** Implements an `Accept()` loop. Every newly connected client is spun off into its own goroutine (`go handleConnection`). This package routes parsed RESP commands to the appropriate underlying subsystems (Store, AOF, Replication).

### `internal/resp` (The Protocol Parser)
- **Role:** Translates raw TCP byte streams into structured Go data types (`resp.Value`), and serializes Go structs back into raw bytes for network transmission.
- **Mechanism:** Uses `bufio.Reader` for efficient, buffered byte inspection. It reads the type identifier byte (`*` for Arrays, `$` for Bulk Strings) and recursively parses lengths and values based on the RESP specification.

### `internal/store` (The Storage Engine)
- **Role:** Maintains the actual key-value data in memory and ensures thread-safe access.
- **Mechanism:** Wraps a standard Go `map[string]string` with a `sync.RWMutex`.
  - **Reads (`GET`):** Use `RLock()`, allowing infinite concurrent readers.
  - **Writes (`SET`, `DEL`):** Use `Lock()`, ensuring exclusive access to prevent data races and map corruption.

### `internal/aof` (The Persistence Manager)
- **Role:** Ensures data survives process termination by logging every mutating operation to disk.
- **Mechanism:** Protects file I/O with a `sync.Mutex`. Every successful `SET` command is converted back to raw RESP bytes and appended to `database.aof`. On startup, the `aof.Read()` function uses the standard `resp` parser to replay the file line-by-line and rebuild the memory state.

### `internal/replication` (The Cluster Manager)
- **Role:** Manages high-availability and read-scaling through Master-Replica syncing.
- **Mechanism:** Maintains a thread-safe list of active TCP connections to followers. 
  - **Full State Transfer:** When a follower sends the `SYNC` command, the leader dumps its entire `.aof` file into the socket.
  - **Live Streaming:** Subsequent `SET` commands are broadcasted to all active follower connections asynchronously.

---

## 3. Request Lifecycles (Data Flow)

### 3.1 The Write Path (`SET`)
1. **Client** sends raw bytes: `*3\r\n$3\r\nSET...`
2. **TCP Handler** streams bytes into the `resp` parser.
3. **Parser** constructs a `resp.Value` AST representing `[SET, key, value]`.
4. **Router** locks the `store` (exclusive write lock) and updates the map.
5. **AOF Manager** locks the file and appends the raw bytes to disk.
6. **Replication Manager** broadcasts the raw bytes to all connected followers.
7. **TCP Handler** sends `+OK\r\n` back to the Client.

### 3.2 The Read Path (`GET`)
1. **Client** sends raw bytes for `GET`.
2. **Parser** extracts the key.
3. **Router** acquires a read-lock (`RLock`) on the `store`.
4. **Store** returns the value (or a boolean `false` if missing).
5. **TCP Handler** formats the value as a RESP Bulk String (`$len\r\nvalue\r\n`) and writes it to the socket.

---

## 4. Key Design Decisions & Trade-offs

### RWMutex vs. Go Channels
While Go's proverb is "share memory by communicating" (channels), FluxKV uses `sync.RWMutex` for the storage layer. 
* *Why:* For a massive concurrent map, standard locking primitives provide drastically lower latency and less garbage collection overhead than passing map operations through channels, especially under read-heavy workloads where `RLock` allows true parallel reads.

### Append-Only File (AOF) vs. Snapshotting (RDB)
FluxKV implements AOF instead of binary point-in-time snapshots.
* *Why:* AOF leverages the existing RESP parser for completely transparent recovery. Replaying text commands is simpler to debug, requires zero custom binary encoding logic, and elegantly solves the initial replication state-transfer problem (just stream the text file).
* *Trade-off:* The AOF file will grow indefinitely. A production iteration would require an "AOF Rewrite" background process to compress redundant keys.

### Goroutine-per-Connection
FluxKV spawns one goroutine per active TCP connection.
* *Why:* This avoids complex `epoll`/`kqueue` asynchronous event loops entirely. The Go runtime scheduler is highly optimized for managing tens of thousands of sleeping goroutines multiplexed over fewer OS threads.