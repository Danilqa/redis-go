package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type App struct {
	storage *map[string]string
}

func main() {
	storage := make(map[string]string)
	app := App{&storage}
	app.run(6379)
}

func (app *App) run(port int16) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to bind to port %d", port))
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
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(val.Array) < 2 {
				conn.Write([]byte(ToSimpleError("ERR wrong number of arguments for 'echo' command")))
				continue
			}
			conn.Write([]byte(ToBulkString(val.Array[1].Str)))
		case "SET":
			if len(val.Array) < 3 {
				conn.Write([]byte(ToSimpleError("ERR wrong number of arguments for 'set' command")))
				continue
			}
			key := val.Array[1]
			value := val.Array[2]
			(*app.storage)[key.Str] = value.Str
			conn.Write([]byte(ToSimpleString("OK")))
		case "GET":
			key := val.Array[1]
			val, ok := (*app.storage)[key.Str]
			if ok {
				conn.Write([]byte(ToBulkString(val)))
			} else {
				conn.Write([]byte(ToNullBulkStrings()))
			}
		default:
			err := fmt.Sprintf("ERR unknown command '%s'", command)
			conn.Write([]byte(ToSimpleError(err)))
		}
	}
}
