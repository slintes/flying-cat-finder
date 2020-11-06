package drone

import (
	"fmt"
	"math"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"

	"flying-cat-finder/api"
)

type telloDrone struct {
	commands chan api.Command
	video    chan<- []byte
	cancel   chan interface{}
	drone    *tello.Driver
}

func New(commands chan api.Command, video chan<- []byte, cancel chan interface{}) *telloDrone {
	if commands == nil || video == nil || cancel == nil {
		panic("tello: args not initialized!")
	}
	return &telloDrone{
		commands: commands,
		video:    video,
		cancel:   cancel,
	}
}

func (t *telloDrone) Start() {
	if t.commands == nil || t.cancel == nil {
		panic("telloDrone fields not initialized!")
	}

	drone := tello.NewDriver("8888")
	t.drone = drone

	work := func() {
		defer func() {
			fmt.Printf("tello: halting\n")
			_ = drone.Halt()
		}()

		go func() {
			select {
			case _, ok := <-t.cancel:
				if !ok {
					fmt.Printf("tello: cancelled!\n")
					close(t.commands)
				}
			}
		}()

		t.handleVideo()
		t.handleCommands()

		fmt.Printf("tello: leaving work func\n")

	}

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{},
		[]gobot.Device{drone},
		work,
	)

	if err := robot.Start(true); err != nil {
		close(t.cancel)
		panic(err)
	}

}

func (t *telloDrone) handleCommands() {
	TurnRate := 0
	ElevationRate := 0
	MoveRate := 0
	SlideRate := 0
	Step := 20

	limit := func(current, step int) int {
		val := float64(current + step)
		val = math.Min(float64(val), 100)
		val = math.Max(float64(val), -100)
		return int(val)
	}

commandLoop:
	for {
		select {
		case command, ok := <-t.commands:

			if !ok {
				fmt.Printf("tello: leaving command loop\n")
				break commandLoop
			}

			fmt.Printf("tello: received command %s\n", command)
			switch command {
			case api.TakeOff:
				if err := t.drone.TakeOff(); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			case api.Land:
				if err := t.drone.Land(); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			case api.Hover:
				TurnRate = 0
				ElevationRate = 0
				MoveRate = 0
				SlideRate = 0
				if err := t.drone.StopLanding(); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
				t.drone.Hover()
			case api.TurnRight:
				TurnRate = limit(TurnRate, Step)
			case api.TurnLeft:
				TurnRate = limit(TurnRate, -Step)
			case api.MoveUp:
				ElevationRate = limit(ElevationRate, Step)
			case api.MoveDown:
				ElevationRate = limit(ElevationRate, -Step)
			case api.MoveRight:
				SlideRate = limit(SlideRate, Step)
			case api.MoveLeft:
				SlideRate = limit(SlideRate, -Step)
			case api.MoveForward:
				MoveRate = limit(MoveRate, Step)
			case api.MoveBackwards:
				MoveRate = limit(MoveRate, -Step)
			}

			fmt.Printf("tello: movement: turn %v, elevate %v, move %v, slide %v", TurnRate, ElevationRate, MoveRate, SlideRate)

			if TurnRate >= 0 {
				if err := t.drone.Clockwise(TurnRate); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			}
			if TurnRate < 0 {
				if err := t.drone.CounterClockwise(-TurnRate); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			}
			if ElevationRate >= 0 {
				if err := t.drone.Up(ElevationRate); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			}
			if ElevationRate < 0 {
				if err := t.drone.Down(-ElevationRate); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			}
			if MoveRate >= 0 {
				if err := t.drone.Forward(MoveRate); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			}
			if MoveRate < 0 {
				if err := t.drone.Backward(-MoveRate); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			}
			if SlideRate >= 0 {
				if err := t.drone.Right(SlideRate); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			}
			if SlideRate < 0 {
				if err := t.drone.Left(-SlideRate); err != nil {
					fmt.Printf("tello: error: %v\n", err)
				}
			}
		}
	}
}

func (t *telloDrone) handleVideo() {

	t.drone.On(tello.ConnectedEvent, func(data interface{}) {
		fmt.Println("Connected")
		t.drone.StartVideo()
		t.drone.SetVideoEncoderRate(4)
		gobot.Every(100*time.Millisecond, func() {
			t.drone.StartVideo()
		})
	})

	t.drone.On(tello.VideoFrameEvent, func(data interface{}) {
		pkt := data.([]byte)
		t.video <- pkt
	})
}
