package cryptox

import "hash/fnv"

func Fnv32HashCode(s string) uint32 {
	f := fnv.New32()
	_, _ = f.Write([]byte(s))
	return f.Sum32()
}

func Fnv32aHashCode(s string) uint32 {
	f := fnv.New32a()
	_, _ = f.Write([]byte(s))
	return f.Sum32()
}

func Fnv64HashCode(s string) uint64 {
	f := fnv.New64()
	_, _ = f.Write([]byte(s))
	return f.Sum64()
}

func Fnv64aHashCode(s string) uint64 {
	f := fnv.New64a()
	_, _ = f.Write([]byte(s))
	return f.Sum64()
}
