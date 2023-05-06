package types

import ipfsref "blitz.build/ipfs-ref"

type DepScope struct {
	Build *ipfsref.IpfsRef[Build]
	Join  *Pathed
	Host  string
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
