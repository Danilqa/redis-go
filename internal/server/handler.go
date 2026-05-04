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

	transactionQueue := []func() string{}
	transactionMode := false
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

		command := strings.ToLower(args.Array[0].Str)
		if command == "discard" {
			if transactionMode == false {
				conn.Write([]byte(resp.SimpleError("ERR DISCARD without MULTI")))
				continue
			}
			transactionMode = false
			transactionQueue = []func() string{}
			conn.Write([]byte(resp.SimpleString("OK")))
			continue
		}
		if command == "multi" {
			transactionMode = true
			conn.Write([]byte(resp.SimpleString("OK")))
			continue
		}
		results := []string{}
		if command == "exec" {
			if transactionMode == false {
				conn.Write([]byte(resp.SimpleError("ERR EXEC without MULTI")))
				continue
			}
			transactionMode = false
			for _, cmd := range transactionQueue {
				results = append(results, cmd())
			}
			conn.Write([]byte(resp.Array(results)))
			continue
		}

		if transactionMode {
			transactionQueue = append(transactionQueue, func() string {
				return srv.dispatch(args)
			})
			conn.Write([]byte(resp.SimpleString("QUEUED")))
		} else {
			reply := srv.dispatch(args)
			conn.Write([]byte(reply))
		}
	}
}

func (srv *Server) dispatch(args resp.Value) string {
	command := strings.ToUpper(args.Array[0].Str)

	switch command {
	case "PING":
		return srv.handlePing()
	case "ECHO":
		return srv.handleEcho(args)
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
		return srv.handleType(args)
	case "XADD":
		return srv.handleXAdd(args)
	case "XRANGE":
		return srv.handleXRange(args)
	case "XREAD":
		return srv.handleXRead(args)
	case "INCR":
		return srv.handleIncr(args)
	default:
		return resp.SimpleError(fmt.Sprintf("ERR unknown command '%s'", command))
	}
}

func (srv *Server) handlePing() string {
	return resp.SimpleString("PONG")
}

func (srv *Server) handleType(args resp.Value) string {
	return resp.SimpleString(srv.storage.GetType(args.Array[1].Str))
}

func (srv *Server) handleEcho(args resp.Value) string {
	if len(args.Array) < 2 {
		return resp.SimpleError("ERR wrong number of arguments for 'echo' command")
	}
	return resp.BulkString(args.Array[1].Str)
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
		return resp.SimpleError("ERR wrong number of arguments for 'xread' command")
	}

	isBlock := strings.ToUpper(args.Array[1].Str) == "BLOCK"
	beforeQuery := 2
	var timeoutMs time.Duration
	if isBlock {
		if len(args.Array) < 6 {
			return resp.SimpleError("ERR wrong number of arguments for 'xread' command")
		}
		d, err := time.ParseDuration(fmt.Sprintf("%sms", args.Array[2].Str))
		if err != nil {
			return resp.SimpleError("ERR invalid timeout argument")
		}
		timeoutMs = d
		beforeQuery = 4
	}

	queryArgs := args.Array[beforeQuery:]
	if len(queryArgs)%2 != 0 {
		return resp.SimpleError("ERR Unbalanced 'xread' list of streams: for each stream key an ID or '$' must be specified.")
	}
	queriesCount := len(queryArgs) / 2

	keys := make([]string, 0, queriesCount)
	fromIDs := make([]string, 0, queriesCount)
	for _, v := range queryArgs[:queriesCount] {
		keys = append(keys, v.Str)
	}
	for _, v := range queryArgs[queriesCount:] {
		fromIDs = append(fromIDs, v.Str)
	}

	results, err := srv.storage.XRead(keys, fromIDs, isBlock, timeoutMs)
	if err != nil {
		return resp.SimpleError(err.Error())
	}
	if results == nil {
		return resp.NullArray()
	}

	r := []string{}
	for _, res := range results {
		if len(res.Entries) == 0 {
			continue
		}
		entrs := []string{}
		for _, entry := range res.Entries {
			fields := make([]string, 0, len(entry.Fields))
			for _, f := range entry.Fields {
				fields = append(fields, resp.BulkString(f))
			}
			entrs = append(entrs, resp.Array([]string{
				resp.BulkString(entry.ID.String()),
				resp.Array(fields),
			}))
		}
		r = append(r, resp.Array([]string{resp.BulkString(res.Key), resp.Array(entrs)}))
	}

	if len(r) == 0 {
		return resp.NullArray()
	}
	return resp.Array(r)
}

func (srv *Server) handleIncr(args resp.Value) string {
	if len(args.Array) < 2 {
		return resp.SimpleError("ERR wrong number of arguments for 'incr' command")
	}
	key := args.Array[1].Str
	newVal, err := srv.storage.Inc(key)
	if err != nil {
		return resp.SimpleError(err.Error())
	}
	return resp.Integer(newVal)
}
