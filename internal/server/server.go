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

func New(s *storage.Storage) *Server {
	info := Info{}
	info["replication"] = []string{"role", "master"}

	return &Server{storage: s, info: &info}
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
