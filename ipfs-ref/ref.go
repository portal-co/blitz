package ipfsref

import (
	"encoding/json"
	"io"

	shell "github.com/ipfs/go-ipfs-api"
)

type IpfsRef[T any] string

func (i IpfsRef[T]) Get(sh *shell.Shell) (res T, err error) {
	c, err := sh.Cat(string(i))
	if err != nil {
		return
	}
	defer c.Close()
	err = json.NewDecoder(c).Decode(&res)
	if err != nil {
		return
	}
	return
}

func Create[T any](sh *shell.Shell, val T) (ref IpfsRef[T], err error) {
	r, w := io.Pipe()
	ch := make(chan struct{})
	go func() {
		json.NewEncoder(w).Encode(val)
		ch <- struct{}{}
	}()
	a, err := sh.Add(r)
	if err != nil {
		return
	}
	<-ch
	return IpfsRef[T](a), nil
}
