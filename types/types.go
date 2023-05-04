package types

type DepScope struct {
	Build *Build
	Join  *Pathed
	Host  string
}
type Build struct {
	Deps map[string]Pathed
	Cmd  []string
}
type Pathed struct {
	Build DepScope
	Path  string
}
