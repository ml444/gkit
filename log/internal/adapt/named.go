package adapt

type Named struct {
	Name string
}

func (n *Named) GetLoggerName() string {
	return n.Name
}

func (n *Named) SetLoggerName(name string) {
	n.Name = name
}
