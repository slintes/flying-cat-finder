package main

import (
	"time"

	"flying-cat-finder/api"
	"flying-cat-finder/pkg/drone"
	"flying-cat-finder/pkg/input"
	"flying-cat-finder/pkg/video"
)

func main() {
	commandsChan := make(chan api.Command, 1)
	videoChan := make(chan []byte, 1)
	cancelChan := make(chan interface{}, 1)

	in := input.New(commandsChan, cancelChan)
	vid := video.New(videoChan, cancelChan)
	dr := drone.New(commandsChan, videoChan, cancelChan)

	go func() {
		dr.Start()
	}()
	go func() {
		vid.Start()
	}()
	go func() {
		in.Start()
	}()

	select {
	case <-cancelChan:
		// give others some time to cleanup
		time.Sleep(1 * time.Second)
	}
}
