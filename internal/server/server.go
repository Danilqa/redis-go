package server

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/storage"
)

type Server struct {
	storage *storage.Storage
	info    *Info
}

type ReplicaOptions struct {
	Host string
	Port int
}

type NewServerOptions struct {
	Storage *storage.Storage
	Replica *ReplicaOptions
}

func New(o NewServerOptions) *Server {
	info := Info{}
	if o.Replica == nil {
		info["replication"] = []string{"role", "master"}
	} else {
		info["replication"] = []string{"role", "slave"}
	}

	return &Server{storage: o.Storage, info: &info}
}

func (srv *Server) Run(port int) {
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
		go srv.handleConnection(conn)
	}
}
