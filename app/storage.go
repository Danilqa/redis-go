package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Storage struct {
	mu      sync.RWMutex
	storage map[string]*StorageValue
}

func (s *Storage) GetValue(args Value) (string, error) {
	key := args.Array[1].Str
	s.mu.RLock()
	item, ok := s.storage[key]
	s.mu.RUnlock()
	if !ok {
		return "", errors.New("Not found")
	}
	if item.IsExpired() {
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

func (s *Storage) SetValue(args Value) (any, error) {
	if len(args.Array) < 3 {
		return nil, errors.New("ERR wrong number of arguments for 'set' command")
	}
	key := args.Array[1].Str
	value := args.Array[2].Str

	var expire *time.Duration

	if len(args.Array) >= 4 {
		expireCommand := strings.ToUpper(args.Array[3].Str)
		if len(args.Array) < 5 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", expireCommand)
		}
		expireValue := args.Array[4].Str

		switch expireCommand {
		case "EX":
			res, err := time.ParseDuration(fmt.Sprintf("%ss", expireValue))
			if err != nil {
				return nil, errors.New("ERR failed to parse expire time")
			}
			expire = &res
		case "PX":
			res, err := time.ParseDuration(fmt.Sprintf("%sms", expireValue))
			if err != nil {
				return nil, errors.New("ERR failed to parse expire time")
			}
			expire = &res
		default:
			return nil, fmt.Errorf("ERR unsupported option '%s'", expireCommand)
		}
	}

	s.mu.Lock()
	s.storage[key] = &StorageValue{
		Value:      value,
		CreatedAt:  time.Now(),
		ExpireInMs: expire,
	}
	s.mu.Unlock()

	return nil, nil
}

func (s *Storage) SetArrayValue(args Value) (int, error) {
	if len(args.Array) < 3 {
		return 0, errors.New("ERR wrong number of arguments for 'rpush' command")
	}
	key := args.Array[1].Str
	value := args.Array[2].Str

	s.mu.Lock()
	existing, ok := s.storage[key]
	if ok {
		existing.Value = append(existing.Value.([]string), value)
		s.mu.Unlock()
		return len(existing.Value.([]string)), nil
	}
	s.storage[key] = &StorageValue{
		Value:     []string{value},
		CreatedAt: time.Now(),
	}
	s.mu.Unlock()

	return 1, nil
}
