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
	"strings"
	"sync"

	"blitz.build/types"
	shell "github.com/ipfs/go-ipfs-api"
	"golang.org/x/sync/errgroup"
)

type BState struct {
	Sh    *shell.Shell
	Alarm *sync.Mutex
}

func (b BState) Log(x string) {
	b.Alarm.Lock()
	defer b.Alarm.Unlock()
	fmt.Fprintln(os.Stderr, x)
}

func hash(x interface{}) []byte {
	s := sha256.New()
	json.NewEncoder(s).Encode(x)
	return s.Sum([]byte{})
}

func buildDep(t types.Pathed, sh BState, stack []string) (p string, err error) {
	var r string
	b := t.Build
	if b.Build != nil {
		var c types.Build
		c, err = b.Build.Get(sh.Sh)
		if err != nil {
			return
		}
		r, err = build(c, sh, stack)
	} else if b.Join != nil {
		var s string
		s, err = buildDep(*b.Join, sh, stack)
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
		r, err = buildDep(x, sh, stack)
	} else {
		r, err = sh.Sh.AddDir(b.Host)
		if err != nil {
			return
		}
	}
	p = r + "/" + t.Path
	return
}

func buildIn(b types.Build, sh BState, t string, stack []string) error {
	defer sh.Log(fmt.Sprintf("Built %s", strings.Join(stack, " > ")))
	// var g errgroup.Group
	// for k, d := range b.Deps {
	// 	k := k
	// 	d := d
	// 	g.Go(func() error {
	// 		p, err := buildDep(d, sh, stack)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		return os.Symlink("/ipfs/"+p, t+"/"+k)
	// 	})
	// 	defer func() {
	// 		os.Remove(t + "/" + k)
	// 	}()
	// }
	// err := g.Wait()
	// if err != nil {
	// 	return err
	// }
	c := exec.Command(b.Cmd[0], b.Cmd[1:]...)
	c.Dir = t
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func build(x types.Build, sh BState, stack []string) (p string, err error) {
	var g errgroup.Group
	var t string
	var mtx sync.Mutex
	key := make(map[string]string)
	for k, d := range x.Deps {
		k := k
		d := d
		g.Go(func() error {
			p, err := buildDep(d, sh, append(stack, x.Name))
			if err != nil {
				return err
			}
			//return os.Symlink("/ipfs/"+p, t+"/"+k)
			mtx.Lock()
			defer mtx.Unlock()
			key[k] = p
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return
	}
	h := "cache/" + base64.URLEncoding.EncodeToString(hash([]interface{}{
		key,
		x,
	}))
	// fmt.Println(h)
	sh.Alarm.Lock()
	if _, err = os.Stat(h); err == nil {
		sh.Alarm.Unlock()
		var x []byte
		x, err = os.ReadFile(h)
		if err != nil {
			return
		}
		p = string(x)
		return
	} else if errors.Is(err, os.ErrNotExist) {
		sh.Alarm.Unlock()
		t, err = os.MkdirTemp("/tmp", "blitz-*")
		if err != nil {
			return
		}
		for k, q := range key {
			k := k
			// fmt.Printf("%s,%s\n", k, p)
			err = os.Symlink("/ipfs/"+q, t+"/"+k)
			if err != nil {
				return
			}
			defer func() {
				os.Remove(t + "/" + k)
			}()
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
			sh.Alarm.Lock()
			defer sh.Alarm.Unlock()
			err = os.WriteFile(h, []byte(d), 0777)
			if err != nil {
				return
			}
			p = d
		}()
		err = buildIn(x, sh, t, append(stack, x.Name))
		return
	}
	sh.Alarm.Unlock()
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
	var alarm sync.Mutex
	s := BState{
		Sh:    sh,
		Alarm: &alarm,
	}
	s.Log("Starting")
	i, err := build(b, s, []string{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Printf("/ipfs/%s\n", i)
}
