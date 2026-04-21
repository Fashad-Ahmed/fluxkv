# FluxKV ⚡️

A high-performance, concurrent, and RESP-compatible distributed key-value store built entirely from scratch in Go.

FluxKV is designed to be a lightweight, Redis-like in-memory database. It implements the Redis Serialization Protocol (RESP), allowing it to interface seamlessly with standard Redis clients (like `redis-cli`) while leveraging Go's powerful concurrency model to handle thousands of simultaneous connections.

## 🚀 Features

- **RESP Compatible:** Implements a custom parser for the Redis Serialization Protocol, meaning standard Redis clients work out of the box.
- **Highly Concurrent:** Utilizes Go's lightweight goroutines to handle multiple TCP client connections without blocking.
- **Thread-Safe Core:** Protects shared memory access using efficient read/write mutexes (`sync.RWMutex`) to ensure data integrity under heavy loads.
- **Zero Dependencies:** The core engine and parser are built using only the Go standard library (`net`, `bufio`, `sync`).

## 🗺️ Project Roadmap

- [x] **Phase 1:** TCP Server & Connection multiplexing
- [x] **Phase 2:** Custom RESP Parser (`BULK`, `ARRAY`, `STRING`)
- [x] **Phase 3:** Thread-safe In-Memory Store (`SET`, `GET`, `DEL`)
- [x] **Phase 4:** Durability via Append-Only File (AOF) persistence
- [x] **Phase 5:** Distributed Consensus & Replication (Clustering)

## 🛠️ Architecture

FluxKV follows a standard, modular Go project layout:

- `cmd/server/`: The entry point of the application, responsible for starting the TCP listener and handling the connection lifecycle.
- `internal/resp/`: Contains the protocol decoding/encoding logic. It translates raw byte streams over the wire into structured Go types.
- `internal/store/`: (WIP) The thread-safe memory engine that processes commands and stores the actual data.

## 💻 Getting Started

### Prerequisites
- Go 1.20+ installed on your machine.
- `redis-cli` (optional, but recommended for testing).

### Installation & Running

1. Clone the repository:
   ```bash
   git clone [https://github.com/Fashad-Ahmed/fluxkv.git](https://github.com/Fashad-Ahmed/fluxkv.git)
   cd fluxkv