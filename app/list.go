package main

import (
	"container/list"
)

type List struct {
	data *list.List
}

func NewList(values []string) *List {
	l := List{}
	l.data = list.New()
	l.Append(values)
	return &l
}

func (l *List) GetRange(start int, stop int) []string {
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

func (l *List) Append(values []string) {
	for _, val := range values {
		l.data.PushBack(val)
	}
}

func (l *List) Prepend(values []string) {
	for _, val := range values {
		l.data.PushFront(val)
	}
}

func (l *List) Len() int {
	return l.data.Len()
}
