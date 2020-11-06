package input

import (
	"flying-cat-finder/api"
	"fmt"
	"github.com/eiannone/keyboard"
)

type Keys struct {
	commands chan<- api.Command
	cancel   chan interface{}
}

func New(commands chan<- api.Command, cancel chan interface{}) *Keys {
	return &Keys{
		commands: commands,
		cancel:   cancel,
	}
}

func (i *Keys) Start() {

	if i.commands == nil || i.cancel == nil {
		panic("Keys fields not initialized!")
	}

	if err := keyboard.Open(); err != nil {
		close(i.cancel)
		panic(err)
	}

	defer func() {
		_ = keyboard.Close()
		close(i.cancel)
	}()

	// interrupt reading keys when program execution is cancelled
	go func() {
		select {
		case _, ok := <-i.cancel:
			if !ok {
				_ = keyboard.Close()
			}
		}
	}()

	fmt.Println("Press ESC to quit")
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			break
		}
		fmt.Printf("  you pressed: rune %q, key %X\r\n", char, key)
		switch key {
		case keyboard.KeyEsc, keyboard.KeyCtrlC:
			close(i.cancel)
		case keyboard.KeyPgup:
			i.commands <- api.TakeOff
		case keyboard.KeyPgdn:
			i.commands <- api.Land
		case keyboard.KeySpace:
			i.commands <- api.Hover
		case keyboard.KeyArrowLeft:
			i.commands <- api.TurnLeft
		case keyboard.KeyArrowRight:
			i.commands <- api.TurnRight
		case keyboard.KeyArrowUp:
			i.commands <- api.MoveUp
		case keyboard.KeyArrowDown:
			i.commands <- api.MoveDown
		}
		switch char {
		case 'a':
			i.commands <- api.MoveLeft
		case 'd':
			i.commands <- api.MoveRight
		case 'w':
			i.commands <- api.MoveForward
		case 's':
			i.commands <- api.MoveBackwards
		}
	}
	fmt.Println("keyboard closed!")
	close(i.cancel)
}
