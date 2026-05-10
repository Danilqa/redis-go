package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/storage"
)

type Server struct {
	storage *storage.Storage
	info    *Info
	replica *ReplicaOptions
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
		info["replication"] = InfoCategory{}
		info["replication"]["role"] = "master"
		info["replication"]["master_replid"] = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		info["replication"]["master_repl_offset"] = "0"
	} else {
		info["replication"] = InfoCategory{}
		info["replication"]["role"] = "slave"
	}

	return &Server{storage: o.Storage, info: &info, replica: o.Replica}
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

func (srv *Server) IsFollower() bool {
	return srv.replica != nil
}

func (srv *Server) ConnectToLeader() {
	conn, err := net.Dial("tcp", net.JoinHostPort(srv.replica.Host, strconv.Itoa(srv.replica.Port)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ping := resp.Array([]string{resp.BulkString("PING")})
	if _, err := conn.Write([]byte(ping)); err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(line)
}
