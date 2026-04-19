package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type App struct {
	Storage *Storage
}

func main() {
	storage := Storage{
		storage: make(map[string]*StorageValue),
		waiters: make(map[string][]chan PopResult),
	}
	app := App{Storage: &storage}
	app.run(6379)
}

func (app *App) run(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go app.handleConnection(conn)
	}
}

func (app *App) handleConnection(conn net.Conn) {
	defer conn.Close()

	parser := NewParser(conn)

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

		switch command := strings.ToUpper(args.Array[0].Str); command {
		case "PING":
			conn.Write([]byte(ToSimpleString("PONG")))
		case "ECHO":
			if len(args.Array) < 2 {
				conn.Write([]byte(ToSimpleError("ERR wrong number of arguments for 'echo' command")))
				continue
			}
			conn.Write([]byte(ToBulkString(args.Array[1].Str)))
		case "SET":
			_, err := app.Storage.SetValue(args)
			if err != nil {
				conn.Write([]byte(ToSimpleError(err.Error())))
				continue
			}

			conn.Write([]byte(ToSimpleString("OK")))
			continue
		case "GET":
			value, err := app.Storage.GetValue(args)
			if err != nil {
				conn.Write([]byte(ToNullBulkStrings()))
				continue
			}

			conn.Write([]byte(ToBulkString(value)))
		case "RPUSH":
			count, err := app.Storage.SetArrayValue(args, false)
			if err != nil {
				conn.Write([]byte(ToSimpleError(err.Error())))
				continue
			}

			conn.Write([]byte(ToInteger(count)))
		case "LPUSH":
			count, err := app.Storage.SetArrayValue(args, true)
			if err != nil {
				conn.Write([]byte(ToSimpleError(err.Error())))
				continue
			}

			conn.Write([]byte(ToInteger(count)))
		case "LRANGE":
			values, err := app.Storage.GetArrayValues(args)
			if err != nil {
				conn.Write([]byte(ToArray([]string{})))
				continue
			}

			arr := make([]string, 0, len(values))
			for _, v := range values {
				arr = append(arr, ToBulkString(v))
			}
			conn.Write([]byte(ToArray(arr)))
		case "LLEN":
			len, err := app.Storage.Len(args)
			if err != nil {
				conn.Write([]byte(ToSimpleError(err.Error())))
				continue
			}
			conn.Write([]byte(ToInteger(len)))
		case "LPOP":
			values, err := app.Storage.Pop(args)
			if err != nil {
				conn.Write([]byte(ToSimpleError(err.Error())))
				continue
			}
			if len(values) == 0 {
				conn.Write([]byte(ToNullBulkStrings()))
				continue
			}
			if len(values) == 1 {
				conn.Write([]byte(ToBulkString(values[0])))
				continue
			}
			arr := make([]string, 0, len(values))
			for _, v := range values {
				arr = append(arr, ToBulkString(v))
			}
			conn.Write([]byte(ToArray(arr)))
		case "BLPOP":
			result, err := app.Storage.BlockPop(args)
			if err != nil {
				conn.Write([]byte(ToSimpleError(err.Error())))
				continue
			}
			if result == nil {
				conn.Write([]byte(ToNullArray()))
			} else {
				conn.Write([]byte(ToArray([]string{
					ToBulkString(result.Key),
					ToBulkString(result.Value),
				})))
			}
		case "TYPE":
			t := app.Storage.GetType(args)
			conn.Write([]byte(ToSimpleString(t)))
		case "XADD":
			id, err := app.Storage.AddStreamValue(args)
			if err != nil {
				conn.Write([]byte(ToSimpleError(err.Error())))
				continue
			}

			conn.Write([]byte(ToSimpleString(id)))
		default:
			err := fmt.Sprintf("ERR unknown command '%s'", command)
			conn.Write([]byte(ToSimpleError(err)))
		}
	}
}
