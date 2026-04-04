package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Storage struct {
	Storage *map[string]StorageValue
}

func (s *Storage) GetValue(args Value) (string, error) {
	key := args.Array[1]
	item, ok := (*s.Storage)[key.Str]
	if ok {
		if item.IsExpired() {
			delete((*s.Storage), key.Str)
			return "", errors.New("Not found")
		}
		return item.Value, nil
	} else {
		return "", errors.New("Not found")
	}
}

func (s *Storage) SetValue(args Value) (any, error) {
	if len(args.Array) < 3 {
		return nil, errors.New("ERR wrong number of arguments for 'set' command")
	}
	key := args.Array[1]
	value := args.Array[2]
	expireCommand := args.Array[3]
	expireValue := args.Array[4]
	if !expireCommand.IsNull && expireValue.IsNull {
		return nil, errors.New(fmt.Sprintf("ERR wrong number of arguments for '%s' command", expireValue.Str))
	}

	var expireInMs time.Duration
	switch strings.ToUpper(expireCommand.Str) {
	case "EX":
		{
			res, err := time.ParseDuration(fmt.Sprintf("%ss", expireValue.Str))
			if err != nil {
				return nil, errors.New("ERR failed to parse expire time")
			}
			expireInMs = res
		}
	case "PX":
		{
			res, err := time.ParseDuration(fmt.Sprintf("%sms", expireValue.Str))
			if err != nil {
				return nil, errors.New("ERR failed to parse expire time")
			}
			expireInMs = res
		}
	}

	(*s.Storage)[key.Str] = StorageValue{
		Value:      value.Str,
		CreatedAt:  time.Now(),
		ExpireInMs: &expireInMs,
	}

	return nil, nil
}
