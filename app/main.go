package main

import (
	"flag"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/storage"
)

func main() {
	portStr := flag.String("port", "6379", "port to listen on")
	flag.Parse()
	srv := server.New(storage.New())
	port, err := strconv.ParseInt(*portStr, 10, 32)
	if err != nil {
		panic("ERR invalid port value")
	}
	srv.Run(int(port))
}
