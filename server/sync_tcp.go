package server

import (
	"io"
	"log"
	"net"

	"github.com/abhinavvsinhaa/redis-go/config"
)

func readCommand(conn net.Conn) (string, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf) // n: number of bytes read
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil // Convert the byte slice to a string and return it
}

func respond(conn net.Conn, cmd string) error {
	_, err := conn.Write([]byte(cmd))
	return err
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
			if err := respond(conn, cmd); err != nil {
				log.Println("Error responding to client:", conn.RemoteAddr(), ", err:", err)
			}
		}
	}
}
