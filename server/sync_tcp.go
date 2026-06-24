package server

import (
	"io"
	"log"
	"net"
	"strings"

	"github.com/abhinavvsinhaa/redis-go/config"
	"github.com/abhinavvsinhaa/redis-go/core"
)

func readCommand(conn net.Conn) (*core.RedisCmd, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf) // n: number of bytes read
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Command: strings.ToUpper(tokens[0]),
		Args:    tokens[1:],
	}, nil
}

func respond(conn net.Conn, cmd *core.RedisCmd) {
	response, err := core.EvalAndRespond(conn, cmd)
	if err != nil {
		conn.Write([]byte(core.EncodeError(err)))
		return
	}
	conn.Write(response)
}

func RunSyncTCPServer() {
	log.Println("Starting Sync TCP Server...")

	var con_clients int = 0

	lsnr, err := net.Listen("tcp", config.Addr())
	if err != nil {
		log.Println("Error starting Sync TCP Server:", err)
		panic(err)
	}

	defer lsnr.Close()

	for {
		conn, err := lsnr.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			panic(err)
		}

		con_clients += 1
		log.Println("Client connected with address:", conn.RemoteAddr(), ", total clients connected:", con_clients)

		for {
			cmd, err := readCommand(conn)
			if err != nil {
				conn.Close()
				con_clients -= 1
				log.Println("Client disconnected with address:", conn.RemoteAddr(), ", total clients connected:", con_clients)

				// Handle EOF error separately to avoid logging it as an error
				// EOF indicates that the client has closed the connection
				if err == io.EOF {
					log.Println("Client closed connection with address:", conn.RemoteAddr())
					break
				}
				log.Println("Error reading command, on client:", conn.RemoteAddr(), ", err:", err)
			}

			log.Println("Received command:", cmd)
			respond(conn, cmd)
		}
	}
}
