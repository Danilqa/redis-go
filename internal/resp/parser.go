package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(line) < 2 {
		return "", errors.New("invalid line")
	}
	return line[:len(line)-2], nil
}

func (p *Parser) readExact(n int) (string, error) {
	buf := make([]byte, n+2)
	_, err := io.ReadFull(p.reader, buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func (p *Parser) Parse() (Value, error) {
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch typeByte {
	case '+':
		return p.parseSimpleString()
	case '-':
		return p.parseError()
	case ':':
		return p.parseInteger()
	case '$':
		return p.parseBulkString()
	case '*':
		return p.parseArray()
	default:
		return Value{}, fmt.Errorf("unknown type byte: %q", typeByte)
	}
}

func (p *Parser) parseSimpleString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Typ: '+', Str: line}, nil
}

func (p *Parser) parseError() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Typ: '-', Str: line}, nil
}

func (p *Parser) parseInteger() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	n, err := strconv.ParseInt(line, 10, 64)
	fmt.Printf("%d", n)
	if err != nil {
		return Value{}, fmt.Errorf("invalid integer: %w", err)
	}
	return Value{Typ: ':', Num: n}, nil
}

func (p *Parser) parseBulkString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid bulk length: %w", err)
	}
	if length == -1 {
		return Value{Typ: '$', IsNull: true}, nil
	}

	data, err := p.readExact(length)
	if err != nil {
		return Value{}, err
	}
	return Value{Typ: '$', Str: data}, nil
}

func (p *Parser) parseArray() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	count, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid array count: %w", err)
	}
	if count == -1 {
		return Value{Typ: '*', IsNull: true}, nil
	}

	items := make([]Value, 0, count)
	for range count {
		val, err := p.Parse()
		if err != nil {
			return Value{}, err
		}
		items = append(items, val)
	}
	return Value{Typ: '*', Array: items}, nil
}
