package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	Typ   string
	Str   string
	Num   int
	Array []Value
}

/**
	Resp is our wrapper around the raw TCP connection
*/
type Resp struct {
	reader *bufio.Reader
}

/**
	NewResp creates a new reader from the raw connection
*/
func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}