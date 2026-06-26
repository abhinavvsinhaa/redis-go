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

// EvalGet retrieves the value associated with the specified key from the store.
// If the key does not exist or has expired, it returns a NilString.
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

// EvalSet sets the value for the specified key in the store, with an optional expiration time.
// The expiration time can be specified in seconds (EX) or milliseconds (PX).
// If the key already exists, its value and expiration time will be updated.
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

// EvalTTL returns the remaining time to live of a key in seconds.
// If the key does not exist, it returns -2.
// If the key exists but has no associated expiration time, it returns -1.
// If the key exists and has an expiration time, it returns the remaining time to live in seconds.
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

// EvalDelete removes the specified keys from the store.
// It returns the number of keys that were successfully deleted.
func EvalDelete(cmd *RedisCmd) []byte {
	if len(cmd.Args) < 1 {
		return []byte(Encode(errors.New("ERR wrong number of arguments for 'del' command")))
	}

	deletedCount := 0
	for _, key := range cmd.Args {
		if _, exists := Get(key); exists {
			Delete(key)
			deletedCount++
		}
	}
	return []byte(Encode(int64(deletedCount)))
}

// EvalExpire sets the expiration time for a key in seconds.
// It returns 1 if the expiration was set successfully, or 0 if the key does not exist.
func EvalExpire(cmd *RedisCmd) []byte {
	if len(cmd.Args) != 2 {
		return []byte(Encode(errors.New("ERR wrong number of arguments for 'expire' command")))
	}

	key := cmd.Args[0]
	expireSeconds, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		return []byte(Encode(errors.New("(error) ERR value is not an integer or out of range")))
	}

	obj, exists := Get(key)
	if !exists {
		return []byte(Encode(int64(0)))
	}

	obj.ExpiresAt = time.Now().UnixMilli() + expireSeconds*1000
	Set(key, obj)

	return []byte(Encode(int64(1)))
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
	case "DEL":
		return EvalDelete(cmd), nil
	case "EXPIRE":
		return EvalExpire(cmd), nil
	default:
		return []byte(Encode(errors.New(fmt.Sprintf("ERR unknown command '%s'", cmd.Command)))), nil
	}
}
