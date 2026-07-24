package crypto

type Mode int

const (
	ModeStorage Mode = iota + 1
	ModeSearchable
)

func (m Mode) Valid() bool {
	return m == ModeStorage || m == ModeSearchable
}
