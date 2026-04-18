package main

import (
	"sync"
)

type Storage struct {
	mu      sync.RWMutex
	storage map[string]*StorageValue
	waiters map[string][]chan PopResult
}

type PopResult struct {
	Key   string
	Value string
}

func (s *Storage) GetType(args Value) string {
	key := args.Array[1].Str
	if v := s.storage[key]; v != nil {
		return v.Type
	} else {
		return "none"
	}
}

func (s *Storage) RemoveWaiter(key string, ch chan PopResult) {
	waiters := s.waiters[key]
	for i, w := range waiters {
		if w == ch {
			s.waiters[key] = append(waiters[:i], waiters[i+1:]...)
			return
		}
	}
}
