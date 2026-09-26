package contracts

type Doable interface {
	Do()
}

type Playable interface {
	Name() string
	Args() []string
}
