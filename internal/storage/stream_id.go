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

func (a streamID) cmp(b streamID) int {
	switch {
	case a.Ms != b.Ms && a.Ms < b.Ms:
		return -1
	case a.Ms != b.Ms:
		return 1
	case a.Seq < b.Seq:
		return -1
	case a.Seq > b.Seq:
		return 1
	}
	return 0
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

func createStreamID(str string) (*streamID, error) {
	msStr, seqStr, ok := strings.Cut(str, "-")
	if !ok {
		return nil, fmt.Errorf("invalid format")
	}
	seq, err := strconv.ParseUint(seqStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid seq in %q: %w", str, err)
	}

	ms, err := strconv.ParseUint(msStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid seq in %q: %w", str, err)
	}

	return &streamID{Ms: ms, Seq: seq}, nil
}

func (s *Storage) createOrderedStreamID(str string, last *streamID) (*streamID, error) {
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

func (id *streamID) String() string {
	ms := strconv.FormatInt(int64(id.Ms), 10)
	seq := strconv.FormatInt(int64(id.Seq), 10)
	return strings.Join([]string{ms, seq}, "-")
}

func createRangeStreamID(str string, defaultBound uint64) (streamID, error) {
	if str == "-" || str == "_" {
		return streamID{Ms: defaultBound, Seq: defaultBound}, nil
	}
	msStr, seqStr, hasSeq := strings.Cut(str, "-")
	ms, err := strconv.ParseUint(msStr, 10, 64)
	if err != nil {
		return streamID{}, fmt.Errorf("invalid ms in %q: %w", str, err)
	}
	seq := defaultBound
	if hasSeq {
		seq, err = strconv.ParseUint(seqStr, 10, 64)
		if err != nil {
			return streamID{}, fmt.Errorf("invalid seq in %q: %w", str, err)
		}
	}
	return streamID{Ms: ms, Seq: seq}, nil
}
