package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (srv *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	parser := resp.NewParser(conn)

	for {
		args, err := parser.Parse()
		if err != nil {
			fmt.Printf("Failed to read buffer %s", err)
			return
		}

		if len(args.Array) == 0 {
			conn.Write([]byte("-ERR empty command\r\n"))
			continue
		}

		reply := srv.dispatch(args)
		conn.Write([]byte(reply))
	}
}

func (srv *Server) dispatch(args resp.Value) string {
	command := strings.ToUpper(args.Array[0].Str)

	switch command {
	case "PING":
		return resp.SimpleString("PONG")
	case "ECHO":
		if len(args.Array) < 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'echo' command")
		}
		return resp.BulkString(args.Array[1].Str)
	case "SET":
		return srv.handleSet(args)
	case "GET":
		return srv.handleGet(args)
	case "RPUSH":
		return srv.handlePush(args, false)
	case "LPUSH":
		return srv.handlePush(args, true)
	case "LRANGE":
		return srv.handleLRange(args)
	case "LLEN":
		return srv.handleLLen(args)
	case "LPOP":
		return srv.handleLPop(args)
	case "BLPOP":
		return srv.handleBLPop(args)
	case "TYPE":
		return resp.SimpleString(srv.storage.GetType(args.Array[1].Str))
	case "XADD":
		return srv.handleXAdd(args)
	case "XRANGE":
		return srv.handleXRange(args)
	case "XREAD":
		return srv.handleXRead(args)
	default:
		return resp.SimpleError(fmt.Sprintf("ERR unknown command '%s'", command))
	}
}

func (srv *Server) handleSet(args resp.Value) string {
	if len(args.Array) < 3 {
		return resp.SimpleError("ERR wrong number of arguments")
	}
	key := args.Array[1].Str
	value := args.Array[2].Str

	var expire *time.Duration
	if len(args.Array) >= 4 {
		expireCommand := strings.ToUpper(args.Array[3].Str)
		if len(args.Array) < 5 {
			return resp.SimpleError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", expireCommand))
		}
		expireValue := args.Array[4].Str

		switch expireCommand {
		case "EX":
			d, err := time.ParseDuration(fmt.Sprintf("%ss", expireValue))
			if err != nil {
				return resp.SimpleError("ERR failed to parse expire time")
			}
			expire = &d
		case "PX":
			d, err := time.ParseDuration(fmt.Sprintf("%sms", expireValue))
			if err != nil {
				return resp.SimpleError("ERR failed to parse expire time")
			}
			expire = &d
		default:
			return resp.SimpleError(fmt.Sprintf("ERR unsupported option '%s'", expireCommand))
		}
	}

	srv.storage.Set(key, value, expire)
	return resp.SimpleString("OK")
}

func (srv *Server) handleGet(args resp.Value) string {
	value, err := srv.storage.Get(args.Array[1].Str)
	if err != nil {
		return resp.NullBulkString()
	}
	return resp.BulkString(value)
}

func (srv *Server) handlePush(args resp.Value, isLeft bool) string {
	if len(args.Array) < 3 {
		return resp.SimpleError("ERR wrong number of arguments")
	}
	key := args.Array[1].Str
	values := make([]string, 0, len(args.Array)-2)
	for _, v := range args.Array[2:] {
		values = append(values, v.Str)
	}
	count := srv.storage.Push(key, values, isLeft)
	return resp.Integer(count)
}

func (srv *Server) handleLRange(args resp.Value) string {
	if len(args.Array) < 4 {
		return resp.Array([]string{})
	}
	key := args.Array[1].Str
	start, err := strconv.ParseInt(args.Array[2].Str, 10, 64)
	if err != nil {
		return resp.Array([]string{})
	}
	stop, err := strconv.ParseInt(args.Array[3].Str, 10, 64)
	if err != nil {
		return resp.Array([]string{})
	}

	values := srv.storage.LRange(key, start, stop)
	arr := make([]string, 0, len(values))
	for _, v := range values {
		arr = append(arr, resp.BulkString(v))
	}
	return resp.Array(arr)
}

func (srv *Server) handleLLen(args resp.Value) string {
	if len(args.Array) < 2 {
		return resp.SimpleError("ERR wrong number of arguments")
	}
	return resp.Integer(srv.storage.LLen(args.Array[1].Str))
}

func (srv *Server) handleLPop(args resp.Value) string {
	if len(args.Array) > 3 || len(args.Array) < 2 {
		return resp.SimpleError("ERR wrong number of arguments")
	}
	key := args.Array[1].Str
	count := 1
	if len(args.Array) == 3 {
		parsedCount, err := strconv.ParseInt(args.Array[2].Str, 10, 64)
		if err != nil {
			return resp.SimpleError("ERR invalid arguments")
		}
		count = int(parsedCount)
	}

	values := srv.storage.LPop(key, count)
	if len(values) == 0 {
		return resp.NullBulkString()
	}
	if len(values) == 1 {
		return resp.BulkString(values[0])
	}
	arr := make([]string, 0, len(values))
	for _, v := range values {
		arr = append(arr, resp.BulkString(v))
	}
	return resp.Array(arr)
}

func (srv *Server) handleBLPop(args resp.Value) string {
	key := args.Array[1].Str
	timeout, err := time.ParseDuration(fmt.Sprintf("%ss", args.Array[2].Str))
	if err != nil {
		return resp.SimpleError("ERR invalid timeout argument")
	}

	result, err := srv.storage.BLPop(key, timeout)
	if err != nil {
		return resp.SimpleError(err.Error())
	}
	if result == nil {
		return resp.NullArray()
	}
	return resp.Array([]string{
		resp.BulkString(result.Key),
		resp.BulkString(result.Value),
	})
}

func (srv *Server) handleXAdd(args resp.Value) string {
	key := args.Array[1].Str
	idStr := args.Array[2].Str
	fields := make([]string, 0, len(args.Array)-3)
	for _, v := range args.Array[3:] {
		fields = append(fields, v.Str)
	}

	id, err := srv.storage.XAdd(key, idStr, fields)
	if err != nil {
		return resp.SimpleError(err.Error())
	}
	return resp.BulkString(id)
}

func (srv *Server) handleXRange(args resp.Value) string {
	if len(args.Array) < 4 {
		return resp.SimpleError("ERR wrong number of arguments for 'xrange' command")
	}
	key := args.Array[1].Str
	from := args.Array[2].Str
	to := args.Array[3].Str
	entries, err := srv.storage.XRange(key, from, to)
	if err != nil {
		return resp.SimpleError(err.Error())
	}
	res := []string{}
	for _, entry := range entries {
		fields := make([]string, 0, len(entry.Fields))
		for _, f := range entry.Fields {
			fields = append(fields, resp.BulkString(f))
		}
		res = append(res, resp.Array([]string{
			resp.BulkString(entry.ID.String()),
			resp.Array(fields),
		}))
	}
	return resp.Array(res)
}

func (srv *Server) handleXRead(args resp.Value) string {
	if len(args.Array) < 4 {
		return resp.SimpleError("ERR wrong number of arguments for 'xrange' command")
	}
	key := args.Array[2].Str
	from := args.Array[3].Str
	entries, err := srv.storage.XRead(key, from)
	if err != nil {
		return resp.SimpleError(err.Error())
	}
	entriesStrings := []string{}
	for _, entry := range entries {
		fields := make([]string, 0, len(entry.Fields))
		for _, f := range entry.Fields {
			fields = append(fields, resp.BulkString(f))
		}
		entriesStrings = append(entriesStrings, resp.Array([]string{
			resp.BulkString(entry.ID.String()),
			resp.Array(fields),
		}))
	}
	respArray := resp.Array(entriesStrings)
	respItemByKey := resp.Array([]string{resp.BulkString(key), respArray})
	return resp.Array([]string{respItemByKey})
}
