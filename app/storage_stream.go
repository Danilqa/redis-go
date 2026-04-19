package main

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type Id struct {
	Ms  uint64
	Seq uint64
}

func (id *Id) ValidateMin() error {
	if id.Ms == 0 && id.Seq == 0 {
		return errors.New("ERR The ID specified in XADD must be greater than 0-0")
	}
	return nil
}

func (new *Id) ValidateOrder(prev *Id) error {
	if new.Ms > prev.Ms {
		return nil
	}
	if new.Ms < prev.Ms {
		return errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}
	if new.Seq > prev.Seq {
		return nil
	} else {
		return errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}
}

func IdFromString(str string) (*Id, error) {
	idPair := strings.Split(str, "-")
	ms, err := strconv.ParseInt(idPair[0], 10, 64)
	if err != nil {
		return nil, err
	}
	seq, err := strconv.ParseInt(idPair[1], 10, 64)
	if err != nil {
		return nil, err
	}
	id := Id{Ms: uint64(ms), Seq: uint64(seq)}
	return &id, nil
}

func (new *Id) toString() string {
	ms := strconv.FormatInt(int64(new.Ms), 10)
	seq := strconv.FormatInt(int64(new.Seq), 10)
	return strings.Join([]string{ms, seq}, "-")
}

type LogEntry struct {
	Id     Id
	Fields []string
}

func (s *Storage) AddStreamValue(args Value) (string, error) {
	key := args.Array[1].Str
	id, err := IdFromString(args.Array[2].Str)
	if err != nil {
		return "", err
	}

	err = id.ValidateMin()
	if err != nil {
		return "", err
	}

	if s.storage[key] == nil {
		s.storage[key] = &StorageValue{Type: "stream", CreatedAt: time.Now(), Value: []LogEntry{}}
	}

	if prev := s.storage[key].Value; prev != nil && len(prev.([]LogEntry)) != 0 {
		entries := s.storage[key].Value.([]LogEntry)
		last := entries[len(entries)-1]
		err := id.ValidateOrder(&last.Id)
		if err != nil {
			return "", err
		}
	}

	fields := Map(args.Array[2:], func(item Value, _ int) string { return item.Str })

	logEntry := LogEntry{Id: *id, Fields: fields}
	s.storage[key].Value = append(s.storage[key].Value.([]LogEntry), logEntry)

	return logEntry.Id.toString(), nil
}
