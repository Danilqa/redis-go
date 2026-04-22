package storage

import (
	"fmt"
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

func (s *Storage) XRange(key string, from string, to string) (*[]*logEntry, error) {
	item := s.storage[key]
	if item == nil {
		return nil, fmt.Errorf("ERR key not found")
	}
	entries, ok := item.Value.([]logEntry)
	if !ok {
		return nil, fmt.Errorf("ERR key not found")
	}
	if len(entries) == 0 {
		return &[]*logEntry{}, nil
	}

	fromId, err := parseSearchStreamId(from)
	if err != nil {
		return nil, fmt.Errorf("ERR invalid from")
	}

	toId, err := parseSearchStreamId(to)
	if err != nil {
		return nil, fmt.Errorf("ERR invalid to")
	}
	result := []*logEntry{}
	for _, v := range entries {
		isLeftMatch := v.ID.Ms == fromId.Ms && (fromId.Seq == nil || *fromId.Seq >= v.ID.Seq)
		isRightMatch := v.ID.Ms == toId.Ms && (toId.Seq == nil || *toId.Seq <= v.ID.Seq)
		if isLeftMatch && isRightMatch {
			result = append(result, &v)
		}
	}

	return &result, nil
}
