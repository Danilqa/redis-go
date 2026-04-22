package storage

import (
	"fmt"
	"math"
	"time"
)

type logEntry struct {
	ID     streamID
	Fields []string
}

func (s *Storage) XAdd(key, idStr string, fields []string) (string, error) {
	lastId := s.getLastId(key)
	id, err := s.createStreamID(idStr, lastId)
	if err != nil {
		return "", err
	}

	if err := id.validateMin(); err != nil {
		return "", err
	}

	if s.storage[key] == nil {
		s.storage[key] = &storageValue{Type: "stream", CreatedAt: time.Now(), Value: []logEntry{}}
	}

	if err := id.validateOrder(lastId); err != nil {
		return "", err
	}

	entry := logEntry{ID: *id, Fields: fields}
	s.storage[key].Value = append(s.storage[key].Value.([]logEntry), entry)

	return entry.ID.String(), nil
}

func (s *Storage) getLastId(key string) *streamID {
	item := s.storage[key]
	if item == nil {
		return nil
	}
	entries, ok := item.Value.([]logEntry)
	if !ok || len(entries) == 0 {
		return nil
	}
	id := entries[len(entries)-1].ID
	return &id
}

func (s *Storage) XRange(key string, from string, to string) ([]logEntry, error) {
	fromID, err := parseRangeStreamID(from, 0)
	if err != nil {
		return nil, fmt.Errorf("ERR invalid from")
	}
	toID, err := parseRangeStreamID(to, math.MaxUint64)
	if err != nil {
		return nil, fmt.Errorf("ERR invalid to")
	}

	item := s.storage[key]
	if item == nil {
		return []logEntry{}, nil
	}
	entries, ok := item.Value.([]logEntry)
	if !ok {
		return []logEntry{}, nil
	}

	result := []logEntry{}
	for _, e := range entries {
		if e.ID.cmp(fromID) >= 0 && e.ID.cmp(toID) <= 0 {
			result = append(result, e)
		}
	}
	return result, nil
}
