package main

import (
	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/storage"
)

func main() {
	srv := server.New(storage.New())
	srv.Run(6379)
}
