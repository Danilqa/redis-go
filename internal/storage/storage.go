package storage

import "sync"

type Storage struct {
	mu      sync.RWMutex
	storage map[string]*storageValue

	listWaiters   map[string][]chan PopResult
	streamWaiters map[string][]*streamWaiter
}

type streamWaiter struct {
	fromID streamID
	ch     chan StreamReadResult
}

type PopResult struct {
	Key   string
	Value string
}

func New() *Storage {
	return &Storage{
		storage:       make(map[string]*storageValue),
		listWaiters:   make(map[string][]chan PopResult),
		streamWaiters: make(map[string][]*streamWaiter),
	}
}

func (s *Storage) GetType(key string) string {
	if v := s.storage[key]; v != nil {
		return v.Type
	}
	return "none"
}
