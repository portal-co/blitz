package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"blitz.build/types"
	shell "github.com/ipfs/go-ipfs-api"
	"golang.org/x/sync/errgroup"
)

type BState struct {
	Sh *shell.Shell
}

func hash(x interface{}) []byte {
	s := sha256.New()
	json.NewEncoder(s).Encode(x)
	return s.Sum([]byte{})
}

func buildDep(t types.Pathed, sh BState) (p string, err error) {
	var r string
	b := t.Build
	if b.Build != nil {
		r, err = build(*b.Build, sh)
	} else if b.Join != nil {
		var s string
		s, err = buildDep(*b.Join, sh)
		if err != nil {
			return
		}
		var c io.ReadCloser
		c, err = sh.Sh.Cat(s)
		if err != nil {
			return
		}
		defer c.Close()
		var x types.Pathed
		err = json.NewDecoder(c).Decode(&x)
		if err != nil {
			return
		}
		r, err = buildDep(x, sh)
	} else {
		r, err = sh.Sh.AddDir(b.Host)
		if err != nil {
			return
		}
	}
	p = r + "/" + t.Path
	return
}

func buildIn(b types.Build, sh BState, t string) error {
	var g errgroup.Group
	for k, d := range b.Deps {
		k := k
		d := d
		g.Go(func() error {
			p, err := buildDep(d, sh)
			if err != nil {
				return err
			}
			return os.Symlink("/ipfs/"+p, t+"/"+k)
		})
		defer func() {
			os.Remove(t + "/" + k)
		}()
	}
	err := g.Wait()
	if err != nil {
		return err
	}
	c := exec.Command(b.Cmd[0], b.Cmd[1:]...)
	c.Dir = t
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func build(x types.Build, sh BState) (p string, err error) {
	h := "cache/" + base64.URLEncoding.EncodeToString(hash(x))
	// fmt.Println(h)
	if _, err = os.Stat(h); err == nil {
		var x []byte
		x, err = os.ReadFile(h)
		if err != nil {
			return
		}
		p = string(x)
	} else if errors.Is(err, os.ErrNotExist) {
		var t string
		t, err = os.MkdirTemp("/tmp", "blitz-*")
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				return
			}
			var d string
			d, err = sh.Sh.AddDir(t)
			if err != nil {
				return
			}
			err = os.RemoveAll(t)
			if err != nil {
				return
			}
			err = os.WriteFile(h, []byte(d), 0777)
			if err != nil {
				return
			}
			p = d
		}()
		err = buildIn(x, sh, t)
		return
	}
	return
}

func main() {
	sh := shell.NewLocalShell()
	var b types.Build
	j := os.Args[1]
	o, err := os.Open(j)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer o.Close()
	err = json.NewDecoder(o).Decode(&b)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	i, err := build(b, BState{
		Sh: sh,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Printf("/ipfs/%s\n", i)
}
