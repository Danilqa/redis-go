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
	port    int
	leader  *LeaderOptions
}

type LeaderOptions struct {
	Host string
	Port int
}

type NewServerOptions struct {
	Port    int
	Storage *storage.Storage
	Leader  *LeaderOptions
}

func New(o NewServerOptions) *Server {
	info := Info{}
	if o.Leader == nil {
		info["replication"] = InfoCategory{}
		info["replication"]["role"] = "master"
		info["replication"]["master_replid"] = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		info["replication"]["master_repl_offset"] = "0"
	} else {
		info["replication"] = InfoCategory{}
		info["replication"]["role"] = "slave"
	}

	return &Server{storage: o.Storage, info: &info, leader: o.Leader, port: o.Port}
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
	return srv.leader != nil
}

func (srv *Server) ConnectToLeader() {
	conn, err := net.Dial("tcp", net.JoinHostPort(srv.leader.Host, strconv.Itoa(srv.leader.Port)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	pingResp := request(&conn, resp.Array([]string{"PING"}))
	println(pingResp)
	request(&conn, resp.Array([]string{"REPLCONF", "listening-port", strconv.Itoa(srv.port)}))
	request(&conn, resp.Array([]string{"REPLCONF", "capa", "psync2"}))
	request(&conn, resp.Array([]string{"PSYNC", "?", "-1"}))
}

func request(conn *net.Conn, command string) string {
	if _, err := (*conn).Write([]byte(command)); err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(*conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return line
}
