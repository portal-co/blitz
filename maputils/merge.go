package maputils

func Merge[K comparable, V any](a, b map[K]V) map[K]V {
	c := make(map[K]V)
	for d, e := range a {
		c[d] = e
	}
	for d, e := range b {
		c[d] = e
	}
	return c
}
