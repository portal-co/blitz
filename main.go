package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	shell "github.com/ipfs/go-ipfs-api"
	"golang.org/x/sync/errgroup"
)

type DepScope struct {
	Build *Build
	Join  *Pathed
}
type Build struct {
	Deps map[string]Pathed
	Cmd  []string
}
type Pathed struct {
	Build DepScope
	Path  string
}

func buildDep(t Pathed, sh *shell.Shell) (p string, err error) {
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
		c, err = sh.Cat(s)
		if err != nil {
			return
		}
		defer c.Close()
		var x Pathed
		err = json.NewDecoder(c).Decode(&x)
		if err != nil {
			return
		}
		r, err = buildDep(x, sh)
	}
	p = r + "/" + t.Path
	return
}

func buildIn(b Build, sh *shell.Shell, t string) error {
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

func build(x Build, sh *shell.Shell) (p string, err error) {
	t, err := os.MkdirTemp("/tmp", "blitz-*")
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			var d string
			d, err = sh.AddDir(t)
			if err == nil {
				err = os.RemoveAll(t)
				if err == nil {
					p = d
				}
			}
		}
	}()
	err = buildIn(x, sh, t)
	return
}

func main() {
	sh := shell.NewLocalShell()
	var b Build
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
	i, err := build(b, sh)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Printf("/ipfs/%s\n", i)
}
