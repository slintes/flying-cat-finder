package input

import (
	"fmt"

	"github.com/eiannone/keyboard"

	"flying-cat-finder/api"
)

type Keys struct {
	commands chan<- api.Command
	cancel   chan interface{}
}

func New(commands chan<- api.Command, cancel chan interface{}) *Keys {
	if commands == nil || cancel == nil {
		panic("keys: args not initialized!")
	}
	return &Keys{
		commands: commands,
		cancel:   cancel,
	}
}

func (k *Keys) Start() {

	if err := keyboard.Open(); err != nil {
		close(k.cancel)
		panic(err)
	}

	defer func() {
		_ = keyboard.Close()
	}()

	// interrupt reading keys when program execution is cancelled
	go func() {
		select {
		case _, ok := <-k.cancel:
			if !ok {
				_ = keyboard.Close()
			}
		}
	}()

	fmt.Println("Press ESC to quit")
keysLoop:
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Printf("keys: error: %v", err)
			close(k.cancel)
			break keysLoop
		}
		fmt.Printf("  keys: you pressed: rune %q, key %X\r\n", char, key)
		switch key {
		case keyboard.KeyEsc, keyboard.KeyCtrlC:
			close(k.cancel)
			break keysLoop
		case keyboard.KeyPgup:
			k.commands <- api.TakeOff
		case keyboard.KeyPgdn:
			k.commands <- api.Land
		case keyboard.KeySpace:
			k.commands <- api.Hover
		case keyboard.KeyArrowLeft:
			k.commands <- api.TurnLeft
		case keyboard.KeyArrowRight:
			k.commands <- api.TurnRight
		case keyboard.KeyArrowUp:
			k.commands <- api.MoveUp
		case keyboard.KeyArrowDown:
			k.commands <- api.MoveDown
		}
		switch char {
		case 'a':
			k.commands <- api.MoveLeft
		case 'd':
			k.commands <- api.MoveRight
		case 'w':
			k.commands <- api.MoveForward
		case 's':
			k.commands <- api.MoveBackwards
		}
	}
	fmt.Println("keys: leaving loop!")
}
