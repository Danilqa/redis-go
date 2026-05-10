package main

import (
	"flag"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/storage"
)

func main() {
	portStr := flag.String("port", "6379", "port to listen on")
	replicaOf := flag.String("replicaof", "", "<PARENT_HOST> <PARENT_PORT>")
	flag.Parse()

	port, err := strconv.ParseInt(*portStr, 10, 32)
	if err != nil {
		panic("ERR invalid port value")
	}

	srv := server.New(server.NewServerOptions{
		Storage: storage.New(),
		Replica: parseReplica(*replicaOf),
	})
	srv.Run(int(port))
}

func parseReplica(s string) *server.ReplicaOptions {
	if s == "" {
		return nil
	}

	parts := strings.Fields(s)
	if len(parts) != 2 {
		panic("ERR --replicaof requires \"<host> <port>\"")
	}

	port, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		panic("ERR invalid replica port value")
	}

	return &server.ReplicaOptions{
		Host: parts[0],
		Port: int(port),
	}
}
