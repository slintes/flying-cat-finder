package main

import (
	"flying-cat-finder/api"
	"flying-cat-finder/pkg/drone"
	"flying-cat-finder/pkg/input"
)

func main() {
	commands := make(chan api.Command, 1)
	cancel := make(chan interface{}, 1)

	in := input.New(commands, cancel)
	dr := drone.New(commands, cancel)

	go func() {
		dr.Start()
	}()
	go func() {
		in.Start()
	}()

	select {
	case <-cancel:
	}
}
