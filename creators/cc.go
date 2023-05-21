package creators

import (
	"fmt"

	"blitz.build/types"
	shell "github.com/ipfs/go-ipfs-api"
)

type CCOpts struct {
	Deps   map[string]types.Pathed
	Srcs   map[string]types.Pathed
	CC     types.Pathed
	ROP    *types.Pathed
	Target string
}

func (c *CCOpts) Emit(sh *shell.Shell) (types.Pathed, error) {
	m := make(map[string]types.Pathed)
	m["cc"] = c.CC
	if c.ROP != nil {
		m["rop"] = *c.ROP
	}
	for k, v := range c.Deps {
		m[k+".o"] = v
	}
	for k, v := range c.Srcs {
		m["src/"+k] = v
	}
	n := "force-exec ./cc $(find . -name '*.o') -Wl,-r -o target.o"
	for k := range c.Srcs {
		n += fmt.Sprintf("-isystem src/%s $(find src/%s - name '*.c')", k, k)
	}
	if c.ROP != nil {
		n += " --target=rop=./rop"
	}
	if c.Target != "" {
		n += " --target=" + c.Target
	}
	return types.B(sh, m, []string{"/bin/sh", "-c", n}, "Compiling CXX object", "./target.o")
}
