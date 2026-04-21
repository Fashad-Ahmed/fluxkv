package store

import "sync"

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]string),
	}
}

func (s *MemoryStore) Set(key, value string) {
	s.mu.Lock()         
	defer s.mu.Unlock() 
	
	s.data[key] = value
}

func (s *MemoryStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	val, exists := s.data[key]
	return val, exists
}

func (s *MemoryStore) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		return true
	}
	return false
}