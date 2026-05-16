package server

import "github.com/codecrafters-io/redis-starter-go/internal/resp"

type transaction struct {
	active bool
	queued []resp.Value
}

func (t *transaction) Active() bool {
	return t.active
}

func (t *transaction) Begin() string {
	if t.active {
		return resp.SimpleError("ERR MULTI calls can not be nested")
	}
	t.active = true
	return resp.SimpleString("OK")
}

func (t *transaction) Queue(args resp.Value) string {
	t.queued = append(t.queued, args)
	return resp.SimpleString("QUEUED")
}

func (t *transaction) Discard() string {
	if !t.active {
		return resp.SimpleError("ERR DISCARD without MULTI")
	}
	t.active = false
	t.queued = nil
	return resp.SimpleString("OK")
}

func (t *transaction) Exec(run func(resp.Value) string) string {
	if !t.active {
		return resp.SimpleError("ERR EXEC without MULTI")
	}
	t.active = false
	queued := t.queued
	t.queued = nil

	results := make([]string, 0, len(queued))
	for _, q := range queued {
		results = append(results, run(q))
	}
	return resp.RawArray(results)
}
