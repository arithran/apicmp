package module

// Diff is a difference between two structures
type Diff struct {
	Path string
	Val1 string
	Val2 string
}

type RawCompare func(one, two []byte) (Diff, bool)
