package api

type Command int

const (
	TakeOff Command = iota
	Land
	Hover
	TurnLeft
	TurnRight
	MoveLeft
	MoveRight
	MoveForward
	MoveBackwards
	MoveUp
	MoveDown
)

func (c Command) String() string {
	return []string{"TakeOff", "Land", "Hover", "TurnLeft", "TurnRight", "MoveLeft", "MoveRight", "MoveForward", "MoveBackwards", "MoveUp", "MoveDown"}[c]
}
