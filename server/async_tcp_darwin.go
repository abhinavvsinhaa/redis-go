//go:build darwin

package server

import (
	"log"
	"net"
	"syscall"
	"time"

	"github.com/abhinavvsinhaa/redis-go/config"
	"github.com/abhinavvsinhaa/redis-go/core"
)

var lastActiveCleanupTime time.Time = time.Now()
var activeCleanupInterval time.Duration = 1 * time.Second

func RunAsyncTCPServer() error {
	log.Println("starting an asynchronous TCP server on", config.Addr())

	max_clients := 20000

	// create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
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

	// Create the kqueue instance
	kqFD, err := syscall.Kqueue()
	if err != nil {
		log.Fatal("Error creating kqueue instance", err)
		return err
	}
	defer syscall.Close(kqFD)

	// Register the server socket for read events
	serverEvent := syscall.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD,
	}
	if _, err = syscall.Kevent(kqFD, []syscall.Kevent_t{serverEvent}, nil, nil); err != nil {
		log.Fatal("Failed to register server file descriptor in kqueue", err)
		return err
	}

	events := make([]syscall.Kevent_t, max_clients)
	conn_clients := 0

	for {
		// Perform active cleanup if the interval has passed, and delete expired keys from the store
		if time.Since(lastActiveCleanupTime) > activeCleanupInterval {
			log.Println("Performing active cleanup...")
			core.ActiveCleanup()
			lastActiveCleanupTime = time.Now()
		}
		// wait for events, with 1s timeout so cleanup can run even when idle
		timeout := syscall.NsecToTimespec(int64(time.Second))
		nevents, err := syscall.Kevent(kqFD, nil, events, &timeout)
		if err != nil {
			log.Println("Error while waiting on kqueue", err)
			continue
		}

		for i := 0; i < nevents; i++ {
			fd := int(events[i].Ident)

			// new incoming connection on server
			if fd == serverFD {
				clientFD, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("Error accepting the connection on fd", clientFD, ", err:", err)
					continue
				}

				conn_clients += 1
				syscall.SetNonblock(clientFD, true)

				// Register the client socket for read events
				clientEvent := syscall.Kevent_t{
					Ident:  uint64(clientFD),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD,
				}
				if _, err = syscall.Kevent(kqFD, []syscall.Kevent_t{clientEvent}, nil, nil); err != nil {
					log.Fatal("Error registering client socket event", err)
				}
			} else {
				conn := core.FdCmd{Fd: fd}
				cmd, err := readCommand(conn)
				if err != nil {
					syscall.Close(fd)
					conn_clients -= 1
					continue
				}

				respond(conn, cmd)
			}
		}
	}
}
