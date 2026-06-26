package core

import (
	"errors"
	"io"
)

func EvalPing(cmd *RedisCmd) []byte {
	if len(cmd.Args) == 0 {
		return []byte(Encode(SimpleString("PONG")))
	} else if len(cmd.Args) == 1 {
		return []byte(Encode(cmd.Args[0]))
	}

	return []byte(Encode(errors.New("ERR wrong number of arguments for 'ping' command")))

}

func EvalAndRespond(conn io.ReadWriter, cmd *RedisCmd) ([]byte, error) {
	switch cmd.Command {
	case "PING":
		return EvalPing(cmd), nil
	default:
		return EvalPing(cmd), nil
	}
}
