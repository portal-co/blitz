package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"os"

	ipfsref "blitz.build/ipfs-ref"
	"blitz.build/types"
	shell "github.com/ipfs/go-ipfs-api"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type BlitzBuild types.Pathed

func (b BlitzBuild) String() string {
	h, _ := b.Hash()
	return fmt.Sprintf("action[%d]", h)
}

func (b BlitzBuild) Type() string {
	return "action"
}

func (b BlitzBuild) Freeze() {

}

func (b BlitzBuild) Truth() starlark.Bool {
	return true
}

func hashFnv(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func (b BlitzBuild) Hash() (uint32, error) {
	h := fnv.New32a()
	json.NewEncoder(h).Encode(b)
	return h.Sum32(), nil
}

type Context struct {
	Shell *shell.Shell
	Funs  map[string]starlark.Callable
}

func (b Context) String() string {
	return "ctx"
}

func (b Context) Type() string {
	return "ctx"
}

func (b Context) Freeze() {

}

func (b Context) Truth() starlark.Bool {
	return true
}

func (b Context) Hash() (uint32, error) {
	return 0, fmt.Errorf("cannot hash a context")
}

func (c Context) AttrNames() []string {
	return []string{"action", "join", "bind"}
}

func (c Context) Attr(name string) (starlark.Value, error) {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if name == "action" {
			var deps *starlark.Dict
			var name string
			var cmd *starlark.List
			var path string
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "deps", &deps, "name", &name, "cmd", &cmd, "path", &path); err != nil {
				return nil, err
			}
			s := make(map[string]types.Pathed)
			for _, k := range deps.Keys() {
				d, _, _ := deps.Get(k)
				s[string(k.(starlark.String))] = types.Pathed(d.(BlitzBuild))
			}
			cmd2 := []string{}
			for i := 0; i < cmd.Len(); i++ {
				v := cmd.Index(i)
				cmd2 = append(cmd2, string(v.(starlark.String)))
			}
			b := types.Build{
				Deps: s,
				Name: name,
				Cmd:  cmd2,
			}
			v, err := ipfsref.Create(c.Shell, b)
			if err != nil {
				return nil, err
			}
			// fmt.Println(string(v))
			return BlitzBuild(types.Pathed{Build: types.DepScope{Build: &v}, Path: path}), nil
		}
		if name == "join" {
			var tgt BlitzBuild
			var path string
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "target", &tgt, "path", &path); err != nil {
				return nil, err
			}
			x := types.Pathed(tgt)
			return BlitzBuild(types.Pathed{Path: path, Build: types.DepScope{Join: &x}}), nil
		}
		// if name == "bind" {
		// 	var tgt BlitzBuild
		// 	var id string
		// 	var run starlark.Callable
		// 	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "target", &tgt, "id", &id, "by", &run); err != nil {
		// 		return nil, err
		// 	}
		// 	hasher := sha256.New()
		// 	hasher.Write([]byte(id))
		// 	hash := base64.URLEncoding.EncodeToString(hasher.Sum([]byte("a")))
		// 	c.Funs[hash] = run

		// }
		return nil, fmt.Errorf("invalid argument")
	}), nil
}

func main() {
	thread := starlark.Thread{
		Name: "root",
	}
	shell := shell.NewLocalShell()
	predeclared := starlark.StringDict{
		"ctx":    Context{Shell: shell},
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
	}
	glob, err := starlark.ExecFile(&thread, os.Args[1], nil, predeclared)
	// fmt.Println("aftr")
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			log.Fatal(evalErr.Backtrace())
		}
		log.Fatal(err)
	}
	// var g errgroup.Group
	for _, name := range glob.Keys() {
		name := name
		v := glob[name]
		if x, ok := v.(BlitzBuild); ok {
			// g.Go(func() error {
			f, err := os.Create("./" + name + ".json")
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			err = json.NewEncoder(f).Encode(types.Build{Cmd: []string{"/bin/echo"}, Name: "Finishing up", Deps: map[string]types.Pathed{"$": types.Pathed(x)}})
			if err != nil {
				log.Fatal(err)
			}
			// })
		}
	}
	// err = g.Wait()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// os.Exit(0)
}
