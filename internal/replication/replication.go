package replication

import (
	"fmt"
	"net"
	"sync"
)

type Manager struct {
	mu        sync.Mutex
	followers []net.Conn
}

func NewManager() *Manager {
	return &Manager{
		followers: make([]net.Conn, 0),
	}
}

func (m *Manager) AddFollower(conn net.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.followers = append(m.followers, conn)
	fmt.Printf("Leader: Follower added (%s). Total followers: %d\n", conn.RemoteAddr().String(), len(m.followers))
}

func (m *Manager) Broadcast(rawCmd []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var activeFollowers []net.Conn

	for _, conn := range m.followers {
		_, err := conn.Write(rawCmd)
		if err != nil {
			fmt.Printf("Leader: Follower disconnected (%s)\n", conn.RemoteAddr().String())
			conn.Close()
		} else {
			activeFollowers = append(activeFollowers, conn)
		}
	}
	
	m.followers = activeFollowers
}