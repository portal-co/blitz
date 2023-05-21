package ipfsref

import (
	"bytes"
	"encoding/json"
	"fmt"

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

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func Create[T any](sh *shell.Shell, val T) (ref IpfsRef[T], err error) {
	e, err := JSONMarshal(val)
	if err != nil {
		return
	}
	a, err := sh.Add(bytes.NewBuffer(e))
	if err != nil {
		return
	}
	fmt.Println("Added")
	return IpfsRef[T](a), nil
}
