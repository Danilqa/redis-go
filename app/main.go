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
	storage := Storage{storage: make(map[string]StorageValue)}
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
		val, err := parser.Parse()
		if err != nil {
			fmt.Printf("Failed to read buffer %s", err)
			return
		}

		if len(val.Array) == 0 {
			conn.Write([]byte("-ERR empty command\r\n"))
			continue
		}

		switch command := strings.ToUpper(val.Array[0].Str); command {
		case "PING":
			conn.Write([]byte(ToSimpleString("PONG")))
		case "ECHO":
			if len(val.Array) < 2 {
				conn.Write([]byte(ToSimpleError("ERR wrong number of arguments for 'echo' command")))
				continue
			}
			conn.Write([]byte(ToBulkString(val.Array[1].Str)))
		case "SET":
			_, err = app.Storage.SetValue(val)
			if err != nil {
				conn.Write([]byte(ToSimpleError(err.Error())))
			} else {
				conn.Write([]byte(ToSimpleString("OK")))
			}
		case "GET":
			value, err := app.Storage.GetValue(val)
			if err != nil {
				conn.Write([]byte(ToNullBulkStrings()))
			} else {
				conn.Write([]byte(ToBulkString(value)))
			}
		default:
			err := fmt.Sprintf("ERR unknown command '%s'", command)
			conn.Write([]byte(ToSimpleError(err)))
		}
	}
}
