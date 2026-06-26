package core

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func EvalPing(cmd *RedisCmd) []byte {
	if len(cmd.Args) == 0 {
		return []byte(Encode(SimpleString("PONG")))
	} else if len(cmd.Args) == 1 {
		return []byte(Encode(cmd.Args[0]))
	}

	return []byte(Encode(errors.New("ERR wrong number of arguments for 'ping' command")))

}

func EvalGet(cmd *RedisCmd) []byte {
	if len(cmd.Args) != 1 {
		return []byte(Encode(errors.New("ERR wrong number of arguments for 'get' command")))
	}

	key := cmd.Args[0]
	obj, exists := Get(key)
	if !exists {
		return []byte(Encode(NilString{}))
	}

	return []byte(Encode(obj.Value))
}

func EvalSet(cmd *RedisCmd) []byte {
	if len(cmd.Args) < 2 {
		return []byte(Encode(errors.New("ERR wrong number of arguments for 'set' command")))
	}

	key, value := cmd.Args[0], cmd.Args[1]
	var expireMillis int64 = -1
	for i := 2; i < len(cmd.Args); i++ {
		arg := strings.ToUpper(cmd.Args[i])
		switch arg {
		case "EX":
			i += 1
			if i >= len(cmd.Args) {
				return []byte(Encode(errors.New("ERR syntax error")))
			}

			expireSeconds, err := strconv.ParseInt(cmd.Args[i], 10, 64)
			if err != nil {
				return []byte(Encode(errors.New("(error) ERR value is not an integer or out of range")))
			}

			expireMillis = expireSeconds * 1000
		case "PX":
			i += 1
			if i >= len(cmd.Args) {
				return []byte(Encode(errors.New("ERR syntax error")))
			}

			px, err := strconv.ParseInt(cmd.Args[i], 10, 64)
			if err != nil {
				return []byte(Encode(errors.New("(error) ERR value is not an integer or out of range")))
			}

			expireMillis = px
		default:
			return []byte(Encode(errors.New("ERR syntax error")))
		}
	}

	Set(key, NewObj(value, expireMillis))
	return []byte(Encode("OK"))
}

func EvalTTL(cmd *RedisCmd) []byte {
	if len(cmd.Args) != 1 {
		return []byte(Encode(errors.New("ERR wrong number of arguments for 'ttl' command")))
	}

	key := cmd.Args[0]
	obj, exists := Get(key)
	if !exists {
		return []byte(Encode(int64(-2)))
	}

	if obj.ExpiresAt == -1 {
		return []byte(Encode(int64(-1)))
	}

	ttlMillis := obj.ExpiresAt - time.Now().UnixMilli()
	if ttlMillis < 0 {
		return []byte(Encode(int64(-2)))
	}

	return []byte(Encode(ttlMillis / 1000))
}

func EvalAndRespond(conn io.ReadWriter, cmd *RedisCmd) ([]byte, error) {
	switch cmd.Command {
	case "PING":
		return EvalPing(cmd), nil
	case "GET":
		return EvalGet(cmd), nil
	case "SET":
		return EvalSet(cmd), nil
	case "TTL":
		return EvalTTL(cmd), nil
	default:
		return []byte(Encode(errors.New(fmt.Sprintf("ERR unknown command '%s'", cmd.Command)))), nil
	}
}
