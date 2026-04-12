package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

var EmptyStrings = []string{}

func (s *Storage) GetArrayValues(args Value) ([]string, error) {
	if len(args.Array) < 4 {
		return EmptyStrings, errors.New("ERR wrong number of arguments")
	}
	key := args.Array[1].Str
	start, err := strconv.ParseInt(args.Array[2].Str, 10, 64)
	if err != nil {
		return EmptyStrings, errors.New("ERR invalid start argument")
	}
	stop, err := strconv.ParseInt(args.Array[3].Str, 10, 64)
	if err != nil {
		return EmptyStrings, errors.New("ERR invalid stop argument")
	}

	s.mu.RLock()
	item, ok := s.storage[key]
	s.mu.RUnlock()

	if !ok {
		return EmptyStrings, nil
	}

	l := item.Value.(*List)
	length := int64(l.Len())

	if start >= length {
		return EmptyStrings, nil
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
		s.notifyWaiters(key, existing)
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
	s.notifyWaiters(key, s.storage[key])
	s.mu.Unlock()

	return len(values), nil
}

func (s *Storage) notifyWaiters(key string, storageRow *StorageValue) {
	if waiters, ok := s.waiters[key]; ok && len(waiters) > 0 {
		ch := waiters[0]
		s.waiters[key] = waiters[1:]
		l := storageRow.Value.(*List)
		vals, _ := l.Pop(1)
		ch <- PopResult{Key: key, Value: vals[0]}
	}
}

func (s *Storage) Pop(args Value) ([]string, error) {
	if len(args.Array) > 3 || len(args.Array) < 2 {
		return EmptyStrings, errors.New("ERR wrong number of arguments")
	}
	key := args.Array[1].Str
	count := 1
	if len(args.Array) == 3 {
		parsedCount, err := strconv.ParseInt(args.Array[2].Str, 10, 64)
		if err != nil {
			return EmptyStrings, errors.New("ERR invalid arguments")
		}
		count = int(parsedCount)
	}

	s.mu.Lock()
	existing, ok := s.storage[key]
	if ok {
		l := existing.Value.(*List)
		removed, err := l.Pop(count)
		if err != nil {
			return EmptyStrings, nil
		}
		s.mu.Unlock()
		return removed, nil
	}

	s.mu.Unlock()
	return EmptyStrings, nil
}

func (s *Storage) BlockPop(args Value) (*PopResult, error) {
	s.mu.Lock()

	key := args.Array[1].Str
	timeout, err := time.ParseDuration(fmt.Sprintf("%ss", args.Array[2].Str))
	if err != nil {
		return nil, errors.New("ERR invalid timeout argument")
	}

	if item, ok := s.storage[key]; ok {
		l := item.Value.(*List)
		if l.Len() > 0 {
			vals, _ := l.Pop(1)
			s.mu.Unlock()
			return &PopResult{Key: key, Value: vals[0]}, nil
		}
	}

	ch := make(chan PopResult, 1)
	s.waiters[key] = append(s.waiters[key], ch)
	s.mu.Unlock()

	if timeout == 0 {
		result := <-ch
		return &result, nil
	}

	select {
	case result := <-ch:
		return &result, nil
	case <-time.After(timeout):
		s.mu.Lock()
		s.RemoveWaiter(key, ch)
		s.mu.Unlock()
		return nil, nil
	}
}
