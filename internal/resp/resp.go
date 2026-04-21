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




/**
	readLine reads byte by byte until it hits \r\n
*/
func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}



/**
	Read parses the incoming byte stream and returns a structured Value
*/
func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.Typ = "array"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	v.Array = make([]Value, len)
	for i := 0; i < len; i++ {
		val, err := r.Read() 
		if err != nil {
			return v, err
		}
		v.Array[i] = val
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.Typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, len)
	r.reader.Read(bulk)
	v.Str = string(bulk)

	r.readLine()

	return v, nil
}

func (v Value) Marshal() []byte {
	switch v.Typ {
	case "array":
		res := []byte(fmt.Sprintf("*%d\r\n", len(v.Array)))
		for _, val := range v.Array {
			res = append(res, val.Marshal()...)
		}
		return res
	case "bulk":
		res := []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.Str), v.Str))
		return res
	case "string":
		res := []byte(fmt.Sprintf("+%s\r\n", v.Str))
		return res
	case "error":
		res := []byte(fmt.Sprintf("-%s\r\n", v.Str))
		return res
	default:
		return []byte{}
	}
}