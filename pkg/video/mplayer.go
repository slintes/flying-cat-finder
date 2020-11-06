package video

import (
	"fmt"
	"os/exec"
)

type mPlayer struct {
	data   <-chan []byte
	cancel chan interface{}
}

func NewPlayer(data <-chan []byte, cancel chan interface{}) *mPlayer {
	if data == nil || cancel == nil {
		panic("mplayer: args not initialized!")
	}
	return &mPlayer{
		data:   data,
		cancel: cancel,
	}
}

func (m *mPlayer) Start() {
	mplayer := exec.Command("mplayer", "-fps", "25", "-")
	mplayerIn, _ := mplayer.StdinPipe()
	if err := mplayer.Start(); err != nil {
		fmt.Printf("mplayer: error: %v", err)
		close(m.cancel)
		return
	}

playerLoop:
	for {
		select {
		case _, ok := <-m.cancel:
			if !ok {
				fmt.Printf("mplayer: cancelled!\n")
				_ = mplayer.Process.Kill()
				break playerLoop
			}
		case data := <-m.data:
			if _, err := mplayerIn.Write(data); err != nil {
				fmt.Println(err)
			}
		}
	}

	fmt.Println("mplayer: leaving loop!")
}
