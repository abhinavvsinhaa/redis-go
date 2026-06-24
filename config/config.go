package config

import (
	"flag"
	"fmt"
)

var Host string = "0.0.0.0"
var Port int = 7379

func Init() {
	flag.StringVar(&Host, "host", "0.0.0.0", "Redis server host")
	flag.IntVar(&Port, "port", 7379, "Redis server port")
	flag.Parse()
}

func Addr() string {
	return fmt.Sprintf("%s:%d", Host, Port)
}
