package server

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type clientSession struct {
	conn   net.Conn
	parser *resp.Parser
	tx     *transaction
}

func newSession(conn net.Conn) *clientSession {
	return &clientSession{
		conn:   conn,
		parser: resp.NewParser(conn),
		tx:     &transaction{},
	}
}
