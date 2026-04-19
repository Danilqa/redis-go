package storage

import "sync"

type Storage struct {
	mu      sync.RWMutex
	storage map[string]*storageValue
	waiters map[string][]chan PopResult
}

type PopResult struct {
	Key   string
	Value string
}

func New() *Storage {
	return &Storage{
		storage: make(map[string]*storageValue),
		waiters: make(map[string][]chan PopResult),
	}
}

func (s *Storage) GetType(key string) string {
	if v := s.storage[key]; v != nil {
		return v.Type
	}
	return "none"
}

func (s *Storage) removeWaiter(key string, ch chan PopResult) {
	waiters := s.waiters[key]
	for i, w := range waiters {
		if w == ch {
			s.waiters[key] = append(waiters[:i], waiters[i+1:]...)
			return
		}
	}
}
