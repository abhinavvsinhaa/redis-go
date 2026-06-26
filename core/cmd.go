package core

import "syscall"

type RedisCmd struct {
	Command string
	Args    []string
}

type FdCmd struct {
	Fd int
}

func (f FdCmd) Write(b []byte) (int, error) {
	return syscall.Write(f.Fd, b)
}

func (f FdCmd) Read(b []byte) (int, error) {
	return syscall.Read(f.Fd, b)
}
