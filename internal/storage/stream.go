package storage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type streamID struct {
	Ms  uint64
	Seq uint64
}

func (id *streamID) validateMin() error {
	if id.Ms == 0 && id.Seq == 0 {
		return errors.New("ERR The ID specified in XADD must be greater than 0-0")
	}
	return nil
}

func (id *streamID) validateOrder(prev *streamID) error {
	if prev == nil {
		return nil
	}
	if id.Ms > prev.Ms {
		return nil
	}
	if id.Ms < prev.Ms {
		return errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}
	if id.Seq > prev.Seq {
		return nil
	}
	return errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
}

func (s *Storage) createStreamID(str string, last *streamID) (*streamID, error) {
	if str == "*" {
		now := uint64(time.Now().UnixMilli())
		if last == nil {
			return &streamID{Ms: now, Seq: 0}, nil
		}
		if last.Ms == now {
			return &streamID{Ms: now, Seq: last.Seq + 1}, nil
		}
	}
	msStr, seqStr, ok := strings.Cut(str, "-")
	if !ok {
		return nil, fmt.Errorf("invalid stream ID %q: expected ms-seq", str)
	}

	ms, err := strconv.ParseUint(msStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ms in %q: %w", str, err)
	}

	var seq uint64 = 0
	switch {
	case seqStr == "*" && last != nil && last.Ms == ms:
		seq = last.Seq + 1
	case seqStr == "*" && ms == 0:
		seq = 1
	case seqStr != "*":
		seq, err = strconv.ParseUint(seqStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid seq in %q: %w", str, err)
		}
	}

	return &streamID{Ms: ms, Seq: seq}, nil
}

func (id *streamID) string() string {
	ms := strconv.FormatInt(int64(id.Ms), 10)
	seq := strconv.FormatInt(int64(id.Seq), 10)
	return strings.Join([]string{ms, seq}, "-")
}

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

	return entry.ID.string(), nil
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
