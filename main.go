package main

import (
	"github.com/abhinavvsinhaa/redis-go/config"
	"github.com/abhinavvsinhaa/redis-go/server"
)

func main() {
	config.Init()
	server.RunSyncTCPServer()
}
