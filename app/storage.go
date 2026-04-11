package main

import (
	"errors"
	"fmt"
	"strconv"
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
		return nil, errors.New("ERR wrong number of arguments")
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

func (s *Storage) GetArrayValues(args Value) ([]string, error) {
	if len(args.Array) < 4 {
		return []string{}, errors.New("ERR wrong number of arguments")
	}
	key := args.Array[1].Str
	start, err := strconv.ParseInt(args.Array[2].Str, 10, 64)
	if err != nil {
		return []string{}, errors.New("ERR invalid start argument")
	}
	stop, err := strconv.ParseInt(args.Array[3].Str, 10, 64)
	if err != nil {
		return []string{}, errors.New("ERR invalid stop argument")
	}

	s.mu.RLock()
	item, ok := s.storage[key]
	s.mu.RUnlock()

	if !ok {
		return []string{}, nil
	}

	l := item.Value.(*List)
	length := int64(l.Len())

	if start >= length {
		return []string{}, nil
	}
	if stop >= length {
		stop = length - 1
	}
	if start < 0 {
		start = max(0, length+start)
	}
	if stop < 0 {
		stop = max(0, length+stop)
	}
	return l.GetRange(int(start), int(stop+1)), nil
}

func (s *Storage) Len(args Value) (int, error) {
	if len(args.Array) < 2 {
		return 0, errors.New("ERR wrong number of arguments")
	}
	key := args.Array[1].Str
	existing, ok := s.storage[key]
	if ok {
		l := existing.Value.(*List)
		return l.Len(), nil
	}

	return 0, nil
}

func (s *Storage) SetArrayValue(args Value, isLeft bool) (int, error) {
	if len(args.Array) < 3 {
		return 0, errors.New("ERR wrong number of arguments")
	}
	key := args.Array[1].Str
	values := []string{}
	for _, v := range args.Array[2:] {
		values = append(values, v.Str)
	}

	s.mu.Lock()
	existing, ok := s.storage[key]
	if ok {
		l := existing.Value.(*List)
		if isLeft {
			l.Prepend(values)
		} else {
			l.Append(values)
		}
		s.mu.Unlock()
		return l.Len(), nil
	}
	l := NewList(nil)
	if isLeft {
		l.Prepend(values)
	} else {
		l.Append(values)
	}
	s.storage[key] = &StorageValue{
		Value:     l,
		CreatedAt: time.Now(),
	}
	s.mu.Unlock()

	return len(values), nil
}
