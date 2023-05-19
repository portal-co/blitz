package types

import (
	ipfsref "blitz.build/ipfs-ref"
	shell "github.com/ipfs/go-ipfs-api"
)

type DepScope struct {
	Build      *ipfsref.IpfsRef[Build]
	Join       *Pathed
	StackEntry string
	FlatMap    *FlatMap
	Host       string
	Worker     *Worker
}
type Build struct {
	Deps map[string]Pathed
	Cmd  []string
	Name string
}
type Pathed struct {
	Build DepScope
	Path  string
}

type FlatMap struct {
	Join Pathed
	As   string
	Push Pathed
}

type Worker struct {
	Code  ipfsref.IpfsRef[Pathed]
	Key   string
	File  string
	Input Pathed
}

func B(sh *shell.Shell, deps map[string]Pathed, cmd []string, name string, path string) (Pathed, error) {
	b, err := ipfsref.Create(sh, Build{
		Deps: deps,
		Cmd:  cmd,
		Name: name,
	})
	return Pathed{
		Path: path,
		Build: DepScope{
			Build: &b,
		},
	}, err
}
func T(x string) Pathed {
	return Pathed{Path: ".", Build: DepScope{StackEntry: x}}
}
func N(x Pathed, n string, g Pathed) Pathed {
	return Pathed{Path: ".", Build: DepScope{FlatMap: &FlatMap{Join: x, As: n, Push: g}}}
}
