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
	s.mu.Lock()
	defer s.mu.Unlock()

	lastId := s.getLastId(key)
	id, err := s.createOrderedStreamID(idStr, lastId)
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

	s.notifyStreamWaiters(key, entry)

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
	fromID, err := createRangeStreamID(from, 0)
	if err != nil {
		return nil, fmt.Errorf("ERR invalid from")
	}
	toID, err := createRangeStreamID(to, math.MaxUint64)
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

type StreamReadResult struct {
	Key     string
	Entries []logEntry
}

func (s *Storage) XRead(
	keys []string,
	fromIDs []string,
	isBlocking bool,
	timeout time.Duration,
) ([]StreamReadResult, error) {
	if len(keys) != len(fromIDs) {
		return nil, fmt.Errorf("ERR Unbalanced 'xread' list of streams: for each stream key an ID or '$' must be specified.")
	}

	s.mu.Lock()

	parsedFromIDs := make([]streamID, 0, len(keys))
	for i, idStr := range fromIDs {
		var id *streamID
		if idStr == "$" {
			id = s.getLastId(keys[i])
		} else {
			createdId, err := createStreamID(idStr)
			if err != nil {
				s.mu.Unlock()
				return nil, fmt.Errorf("ERR invalid from")
			}
			id = createdId
		}

		parsedFromIDs = append(parsedFromIDs, *id)
	}

	results := s.collectStreamEntries(keys, parsedFromIDs)

	hasEntries := false
	for _, r := range results {
		if len(r.Entries) > 0 {
			hasEntries = true
			break
		}
	}

	if !isBlocking || hasEntries {
		s.mu.Unlock()
		return results, nil
	}

	ch := make(chan StreamReadResult, len(keys))
	waiters := make([]*streamWaiter, len(keys))
	for i, key := range keys {
		w := &streamWaiter{fromID: parsedFromIDs[i], ch: ch}
		waiters[i] = w
		s.streamWaiters[key] = append(s.streamWaiters[key], w)
	}
	s.mu.Unlock()

	var got *StreamReadResult
	if timeout == 0 {
		r := <-ch
		got = &r
	} else {
		select {
		case r := <-ch:
			got = &r
		case <-time.After(timeout):
		}
	}

	s.mu.Lock()
	for i, w := range waiters {
		s.removeStreamWaiter(keys[i], w)
	}
	s.mu.Unlock()

	if got == nil {
		return nil, nil
	}
	return []StreamReadResult{*got}, nil
}

func (s *Storage) collectStreamEntries(keys []string, fromIDs []streamID) []StreamReadResult {
	results := make([]StreamReadResult, 0, len(keys))
	for i, key := range keys {
		result := StreamReadResult{Key: key, Entries: []logEntry{}}
		if item := s.storage[key]; item != nil {
			if entries, ok := item.Value.([]logEntry); ok {
				for _, e := range entries {
					if e.ID.cmp(fromIDs[i]) > 0 {
						result.Entries = append(result.Entries, e)
					}
				}
			}
		}
		results = append(results, result)
	}
	return results
}

func (s *Storage) notifyStreamWaiters(key string, entry logEntry) {
	waiters := s.streamWaiters[key]
	if len(waiters) == 0 {
		return
	}
	remaining := waiters[:0]
	for _, w := range waiters {
		if entry.ID.cmp(w.fromID) > 0 {
			select {
			case w.ch <- StreamReadResult{Key: key, Entries: []logEntry{entry}}:
			default:
			}
		} else {
			remaining = append(remaining, w)
		}
	}
	s.streamWaiters[key] = remaining
}

func (s *Storage) removeStreamWaiter(key string, target *streamWaiter) {
	waiters := s.streamWaiters[key]
	for i, w := range waiters {
		if w == target {
			s.streamWaiters[key] = append(waiters[:i], waiters[i+1:]...)
			return
		}
	}
}
