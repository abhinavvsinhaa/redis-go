package core

import (
	"errors"
)

func readLength(data []byte) (int, int) {
	len, pos := 0, 0
	for pos = range data {
		c := data[pos]
		if c == '\r' {
			break
		}

		len = len*10 + int(c-'0')
	}

	return len, pos + 2
}

func readSimpleString(data []byte) (string, int, error) {
	// Simple Strings are prefixed with a '+' character and terminated by '\r\n'
	pos := 1
	for data[pos] != '\r' {
		pos++
	}
	return string(data[1:pos]), pos + 2, nil
}

func readError(data []byte) (string, int, error) {
	// Errors are prefixed with a '-' character and terminated by '\r\n'
	return readSimpleString(data)
}

func readInt64(data []byte) (int64, int, error) {
	// Integers are prefixed with a ':' character and terminated by '\r\n'
	// :1279\r\n
	pos := 1
	var value int64 = 0
	for ; data[pos] != '\r'; pos++ {
		// Convert the byte to an integer and accumulate it into value
		value = value*10 + int64(data[pos]-'0')
	}
	return value, pos + 2, nil
}

func readBulkString(data []byte) (string, int, error) {
	// Bulk Strings are prefixed with a '$' character, followed by the length of the string, and terminated by '\r\n'
	// $62\r\nfoobar...\r\n
	var pos = 1

	// Read the length of the bulk string
	len, delta := readLength(data[pos:])
	pos += delta

	return string(data[pos : pos+len]), pos + len + 2, nil
}

func readArray(data []byte) (interface{}, int, error) {
	// Arrays are prefixed with a '*' character, followed by the number of elements in the array, and terminated by '\r\n'
	// *2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
	var pos = 1

	// Read the number of elements in the array
	numElements, delta := readLength(data[pos:])
	pos += delta

	elements := make([]interface{}, numElements)
	for i := range elements {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elements[i] = elem
		pos += delta
	}
	return elements, pos, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {
	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInt64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	default:
		return nil, 0, errors.New("unknown RESP type")
	}
}

func Decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	value, _, err := DecodeOne(data)
	return value, err
}
