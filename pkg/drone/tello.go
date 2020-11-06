package drone

import (
	"flying-cat-finder/api"
	"fmt"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
	"math"
)

type Tello struct {
	commands <-chan api.Command
	cancel   chan interface{}
}

func New(commands chan api.Command, cancel chan interface{}) *Tello {
	return &Tello{
		commands: commands,
		cancel:   cancel,
	}
}

func (t *Tello) Start() {
	if t.commands == nil || t.cancel == nil {
		panic("Tello fields not initialized!")
	}

	drone := tello.NewDriver("8888")

	work := func() {
		defer func() {
			fmt.Printf("tello: halt\n")
			_ = drone.Halt()
		}()

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
		for {
			select {
			case _, ok := <-t.cancel:
				if !ok {
					fmt.Printf("tello: execution cancelled!\n")
					break
				}
			case command := <-t.commands:
				fmt.Printf("tello: received command %s\n", command)
				switch command {
				case api.TakeOff:
					if err := drone.TakeOff(); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				case api.Land:
					if err := drone.Land(); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				case api.Hover:
					TurnRate = 0
					ElevationRate = 0
					MoveRate = 0
					SlideRate = 0
					if err := drone.StopLanding(); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
					drone.Hover()
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
					if err := drone.Clockwise(TurnRate); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				}
				if TurnRate < 0 {
					if err := drone.CounterClockwise(-TurnRate); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				}
				if ElevationRate >= 0 {
					if err := drone.Up(ElevationRate); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				}
				if ElevationRate < 0 {
					if err := drone.Down(-ElevationRate); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				}
				if MoveRate >= 0 {
					if err := drone.Forward(MoveRate); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				}
				if MoveRate < 0 {
					if err := drone.Backward(-MoveRate); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				}
				if SlideRate >= 0 {
					if err := drone.Right(SlideRate); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				}
				if SlideRate < 0 {
					if err := drone.Left(-SlideRate); err != nil {
						fmt.Printf("tello: error: %v\n", err)
					}
				}
			}
		}
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
