package main

import (
	"strconv"
	"strings"
	"time"
)

type Id struct {
	Ms  uint64
	Seq uint64
}

func (id *Id) toString() string {
	ms := strconv.FormatInt(int64(id.Ms), 10)
	seq := strconv.FormatInt(int64(id.Seq), 10)
	return strings.Join([]string{ms, seq}, "-")
}

type LogEntry struct {
	Id     Id
	Fields []string
}

func (s *Storage) AddStreamValue(args Value) (string, error) {
	key := args.Array[1].Str
	idPair := strings.Split(args.Array[2].Str, "-")

	ms, err := strconv.ParseInt(idPair[0], 10, 64)
	if err != nil {
		return "", err
	}

	seq, err := strconv.ParseInt(idPair[1], 10, 64)
	if err != nil {
		return "", err
	}

	id := Id{Ms: uint64(ms), Seq: uint64(seq)}
	fields := Map(args.Array[2:], func(item Value, _ int) string { return item.Str })

	logEntry := LogEntry{Id: id, Fields: fields}
	if s.storage[key] == nil {
		s.storage[key] = &StorageValue{Type: "stream", CreatedAt: time.Now(), Value: []LogEntry{}}
	}
	s.storage[key].Value = append(s.storage[key].Value.([]LogEntry), logEntry)

	return logEntry.Id.toString(), nil
}
