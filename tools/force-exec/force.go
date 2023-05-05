package main

import (
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/justincormack/go-memfd/msyscall"
)

func main() {
	fd, err := msyscall.MemfdCreate(os.Args[1], msyscall.MFD_CLOEXEC)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	l := os.NewFile(fd, os.Args[1])
	i, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer i.Close()
	_, err = io.Copy(l, i)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for syscall.Exec(fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), l.Fd()), os.Args[1:], os.Environ()) != nil {
	}
}
