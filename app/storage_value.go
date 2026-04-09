package main

import "time"

type StorageValue struct {
	Value      any
	CreatedAt  time.Time
	ExpireInMs *time.Duration
}

func (sv *StorageValue) IsExpired() bool {
	if sv.ExpireInMs == nil {
		return false
	}
	return time.Now().After(sv.CreatedAt.Add(*sv.ExpireInMs))
}
