package storage

import (
	"errors"
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

func parseStreamID(str string) (*streamID, error) {
	pair := strings.Split(str, "-")
	ms, err := strconv.ParseInt(pair[0], 10, 64)
	if err != nil {
		return nil, err
	}
	seq, err := strconv.ParseInt(pair[1], 10, 64)
	if err != nil {
		return nil, err
	}
	return &streamID{Ms: uint64(ms), Seq: uint64(seq)}, nil
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
	id, err := parseStreamID(idStr)
	if err != nil {
		return "", err
	}

	if err := id.validateMin(); err != nil {
		return "", err
	}

	if s.storage[key] == nil {
		s.storage[key] = &storageValue{Type: "stream", CreatedAt: time.Now(), Value: []logEntry{}}
	}

	if prev := s.storage[key].Value; prev != nil && len(prev.([]logEntry)) != 0 {
		entries := s.storage[key].Value.([]logEntry)
		last := entries[len(entries)-1]
		if err := id.validateOrder(&last.ID); err != nil {
			return "", err
		}
	}

	entry := logEntry{ID: *id, Fields: fields}
	s.storage[key].Value = append(s.storage[key].Value.([]logEntry), entry)

	return entry.ID.string(), nil
}
