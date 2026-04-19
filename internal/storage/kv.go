package storage

import (
	"errors"
	"time"
)

func (s *Storage) Get(key string) (string, error) {
	s.mu.RLock()
	item, ok := s.storage[key]
	s.mu.RUnlock()
	if !ok {
		return "", errors.New("Not found")
	}
	if item.isExpired() {
		s.mu.Lock()
		delete(s.storage, key)
		s.mu.Unlock()
		return "", errors.New("Not found")
	}
	str, ok := item.Value.(string)
	if !ok {
		return "", errors.New("ERR value is not a string")
	}
	return str, nil
}

func (s *Storage) Set(key, value string, expire *time.Duration) {
	s.mu.Lock()
	s.storage[key] = &storageValue{
		Value:      value,
		CreatedAt:  time.Now(),
		Type:       "string",
		ExpireInMs: expire,
	}
	s.mu.Unlock()
}
