package storage

import "time"

type storageValue struct {
	Value      any
	Type       string
	CreatedAt  time.Time
	ExpireInMs *time.Duration
}

func (sv *storageValue) isExpired() bool {
	if sv.ExpireInMs == nil {
		return false
	}
	return time.Now().After(sv.CreatedAt.Add(*sv.ExpireInMs))
}
