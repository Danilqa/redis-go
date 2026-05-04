package storage

import (
	"errors"
	"strconv"
	"time"
)

func (s *Storage) Inc(key string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.storage[key]
	if !ok {
		s.storage[key] = &storageValue{
			Value:     "1",
			Type:      "string",
			CreatedAt: time.Now(),
		}
		return 1, nil
	}
	if item.Type != "string" {
		return 0, errors.New("ERR value is not an integer or out of range")
	}

	str, ok := item.Value.(string)
	if !ok {
		return 0, errors.New("ERR value is not an integer or out of range")
	}
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, errors.New("ERR value is not an integer or out of range")
	}

	next := int(num + 1)
	item.Value = strconv.Itoa(next)
	return next, nil
}
