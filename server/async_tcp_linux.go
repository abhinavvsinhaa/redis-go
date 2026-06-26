//go:build linux

package server

import (
	"log"
	"net"
	"syscall"

	"github.com/abhinavvsinhaa/redis-go/config"
	"github.com/abhinavvsinhaa/redis-go/core"
)

func RunAsyncTCPServer() error {
	log.Println("starting an asynchronous TCP server on", config.Addr())

	max_clients := 20000

	// create EPOLL event objects to hold events
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	// create a socket
	// AF_INET -> IPv4 socket, SOCK_STREAM -> keep the tcp connection open, even though we have received the first connection
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Fatal("Error creating server socket", err)
		return err
	}
	defer syscall.Close(serverFD)

	// set the socket to operate in a non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		log.Fatal("Error while setting the socket to operate in a non-blocking mode", err)
		return err
	}

	ip4 := net.ParseIP(config.Host).To4()
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		log.Fatal("Error binding the file descriptor to the given address", err)
		return err
	}

	if err = syscall.Listen(serverFD, max_clients); err != nil {
		log.Fatal("Error listening to the server", err)
		return err
	}

	// Create the epoll instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal("Error creating epoll instance", err)
		return err
	}
	defer syscall.Close(epollFD)

	// specify the events we want to get hints about, and on which socket
	var serverEpollEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// Listen to read events on the server itself
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &serverEpollEvent); err != nil {
		log.Fatal("Failed to register server file descriptor in epoll", err)
		return err
	}

	conn_clients := 0
	for {
		// see if any FD is ready for IO
		nevents, e := syscall.EpollWait(epollFD, events[:], -1)
		if e != nil {
			log.Println("Error while waiting on epoll", e)
			continue
		}

		for i := range nevents {
			// new incoming connection on server
			if events[i].Fd == int32(serverFD) {
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("Error accepting the connection on fd", fd, ", err:", err)
					continue
				}

				conn_clients += 1
				syscall.SetNonblock(fd, true)

				var clientSocketEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}
				if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &clientSocketEvent); err != nil {
					log.Fatal("Error registering client socket event", err)
				}
			} else {
				conn := core.FdCmd{Fd: int(events[i].Fd)}
				cmd, err := readCommand(conn)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					conn_clients -= 1
					continue
				}

				respond(conn, cmd)
			}
		}

	}

}
