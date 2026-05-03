package storage

import (
	"container/list"
	"errors"
	"time"
)

type linkedList struct {
	data *list.List
}

func newList() *linkedList {
	return &linkedList{data: list.New()}
}

func (l *linkedList) getRange(start int, stop int) []string {
	result := []string{}
	curr := l.data.Front()
	i := 0
	for curr != nil && i < stop {
		if i >= start {
			result = append(result, curr.Value.(string))
		}
		curr = curr.Next()
		i++
	}
	return result
}

func (l *linkedList) append(values []string) {
	for _, val := range values {
		l.data.PushBack(val)
	}
}

func (l *linkedList) prepend(values []string) {
	for _, val := range values {
		l.data.PushFront(val)
	}
}

func (l *linkedList) len() int {
	return l.data.Len()
}

func (l *linkedList) pop(n int) ([]string, error) {
	if l.data.Len() == 0 {
		return []string{}, errors.New("Empty list")
	}
	removed := []string{}
	for range n {
		front := l.data.Front()
		l.data.Remove(front)
		removed = append(removed, front.Value.(string))
	}
	return removed, nil
}

func (s *Storage) LRange(key string, start, stop int64) []string {
	s.mu.RLock()
	item, ok := s.storage[key]
	s.mu.RUnlock()

	if !ok {
		return []string{}
	}

	l := item.Value.(*linkedList)
	length := int64(l.len())

	if start >= length {
		return []string{}
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
	return l.getRange(int(start), int(stop+1))
}

func (s *Storage) LLen(key string) int {
	existing, ok := s.storage[key]
	if ok {
		l := existing.Value.(*linkedList)
		return l.len()
	}
	return 0
}

func (s *Storage) Push(key string, values []string, isLeft bool) int {
	s.mu.Lock()
	existing, ok := s.storage[key]
	if ok {
		l := existing.Value.(*linkedList)
		if isLeft {
			l.prepend(values)
		} else {
			l.append(values)
		}
		s.notifyWaiters(key, existing)
		s.mu.Unlock()
		return l.len()
	}
	l := newList()
	if isLeft {
		l.prepend(values)
	} else {
		l.append(values)
	}
	s.storage[key] = &storageValue{
		Value:     l,
		CreatedAt: time.Now(),
		Type:      "list",
	}
	s.notifyWaiters(key, s.storage[key])
	s.mu.Unlock()

	return len(values)
}

func (s *Storage) notifyWaiters(key string, row *storageValue) {
	if waiters, ok := s.listWaiters[key]; ok && len(waiters) > 0 {
		ch := waiters[0]
		s.listWaiters[key] = waiters[1:]
		l := row.Value.(*linkedList)
		vals, _ := l.pop(1)
		ch <- PopResult{Key: key, Value: vals[0]}
	}
}

func (s *Storage) LPop(key string, count int) []string {
	s.mu.Lock()
	existing, ok := s.storage[key]
	if ok {
		l := existing.Value.(*linkedList)
		removed, err := l.pop(count)
		if err != nil {
			s.mu.Unlock()
			return []string{}
		}
		s.mu.Unlock()
		return removed
	}

	s.mu.Unlock()
	return []string{}
}

func (s *Storage) BLPop(key string, timeout time.Duration) (*PopResult, error) {
	s.mu.Lock()

	if item, ok := s.storage[key]; ok {
		l := item.Value.(*linkedList)
		if l.len() > 0 {
			vals, _ := l.pop(1)
			s.mu.Unlock()
			return &PopResult{Key: key, Value: vals[0]}, nil
		}
	}

	ch := make(chan PopResult, 1)
	s.listWaiters[key] = append(s.listWaiters[key], ch)
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
		s.removeListWaiter(key, ch)
		s.mu.Unlock()
		return nil, nil
	}
}

func (s *Storage) removeListWaiter(key string, ch chan PopResult) {
	waiters := s.listWaiters[key]
	for i, w := range waiters {
		if w == ch {
			s.listWaiters[key] = append(waiters[:i], waiters[i+1:]...)
			return
		}
	}
}
